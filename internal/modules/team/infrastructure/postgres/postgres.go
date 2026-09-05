// internal/modules/team/infrastructure/postgres/repository.go

package postgres

import (
    "context"
    "errors"
    "fmt"
    "time"

    "gorm.io/gorm"

    "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/team/teamdomain"
)

// TeamRepository implements teamdomain.Repository using PostgreSQL
type TeamRepository struct {
    db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) teamdomain.Repository {
    return &TeamRepository{db: db}
}

// ============================================================
// TEAM OPERATIONS
// ============================================================

func (r *TeamRepository) CreateTeam(ctx context.Context, team *teamdomain.Team) error {
    model := TeamModel{}
    model.FromDomain(team)
    return r.db.WithContext(ctx).Create(&model).Error
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, id string) (*teamdomain.Team, error) {
    var model TeamModel
    if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get team: %w", err)
    }
    return model.ToDomain(), nil
}

func (r *TeamRepository) GetTeamBySlug(ctx context.Context, slug string) (*teamdomain.Team, error) {
    var model TeamModel
    if err := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get team by slug: %w", err)
    }
    return model.ToDomain(), nil
}

func (r *TeamRepository) GetTeamsByUserID(ctx context.Context, userID string) ([]*teamdomain.Team, error) {
    var models []TeamModel
    if err := r.db.WithContext(ctx).
        Joins("JOIN team_members ON team_members.team_id = teams.id").
        Where("team_members.user_id = ? AND team_members.deleted_at IS NULL", userID).
        Find(&models).Error; err != nil {
        return nil, fmt.Errorf("failed to get teams: %w", err)
    }

    teams := make([]*teamdomain.Team, len(models))
    for i, model := range models {
        teams[i] = model.ToDomain()
    }
    return teams, nil
}

func (r *TeamRepository) UpdateTeam(ctx context.Context, team *teamdomain.Team) error {
    model := TeamModel{}
    model.FromDomain(team)
    return r.db.WithContext(ctx).Save(&model).Error
}

func (r *TeamRepository) DeleteTeam(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Model(&TeamModel{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// ============================================================
// MEMBER OPERATIONS
// ============================================================

func (r *TeamRepository) CreateMember(ctx context.Context, member *teamdomain.Member) error {
    model := MemberModel{}
    model.FromDomain(member)
    return r.db.WithContext(ctx).Create(&model).Error
}

func (r *TeamRepository) GetMemberByID(ctx context.Context, id string) (*teamdomain.Member, error) {
    var model MemberModel
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get member: %w", err)
    }
    return model.ToDomain(), nil
}

func (r *TeamRepository) GetMemberByTeamAndUser(ctx context.Context, teamID, userID string) (*teamdomain.Member, error) {
    var model MemberModel
    if err := r.db.WithContext(ctx).
        Where("team_id = ? AND user_id = ?", teamID, userID).
        First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get member: %w", err)
    }
    return model.ToDomain(), nil
}

func (r *TeamRepository) GetMembersByTeam(ctx context.Context, teamID string, filters teamdomain.ListMembersFilters) ([]*teamdomain.Member, int64, error) {
    query := r.db.WithContext(ctx).Model(&MemberModel{}).Where("team_id = ?", teamID)

    // Apply filters
    if filters.Role != "" {
        query = query.Where("role = ?", filters.Role)
    }

    // Count total
    var total int64
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count members: %w", err)
    }

    // Pagination
    if filters.Limit > 0 {
        query = query.Limit(filters.Limit)
        if filters.Offset > 0 {
            query = query.Offset(filters.Offset)
        }
    }

    var models []MemberModel
    if err := query.Find(&models).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to get members: %w", err)
    }

    members := make([]*teamdomain.Member, len(models))
    for i, model := range models {
        members[i] = model.ToDomain()
    }
    return members, total, nil
}

func (r *TeamRepository) GetMembersByUser(ctx context.Context, userID string) ([]*teamdomain.Member, error) {
    var models []MemberModel
    if err := r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID).Find(&models).Error; err != nil {
        return nil, fmt.Errorf("failed to get memberships: %w", err)
    }

    members := make([]*teamdomain.Member, len(models))
    for i, model := range models {
        members[i] = model.ToDomain()
    }
    return members, nil
}

func (r *TeamRepository) UpdateMember(ctx context.Context, member *teamdomain.Member) error {
    model := MemberModel{}
    model.FromDomain(member)
    return r.db.WithContext(ctx).Save(&model).Error
}

func (r *TeamRepository) DeleteMember(ctx context.Context, teamID, userID string) error {
    return r.db.WithContext(ctx).
        Model(&MemberModel{}).
        Where("team_id = ? AND user_id = ?", teamID, userID).
        Update("deleted_at", time.Now()).Error
}

func (r *TeamRepository) CountMembersByTeam(ctx context.Context, teamID string) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).Model(&MemberModel{}).Where("team_id = ? AND deleted_at IS NULL", teamID).Count(&count).Error; err != nil {
        return 0, fmt.Errorf("failed to count members: %w", err)
    }
    return count, nil
}

// ============================================================
// INVITATION OPERATIONS
// ============================================================

func (r *TeamRepository) CreateInvitation(ctx context.Context, invitation *teamdomain.Invitation) error {
    model := InvitationModel{}
    model.FromDomain(invitation)
    return r.db.WithContext(ctx).Create(&model).Error
}

func (r *TeamRepository) GetInvitationByID(ctx context.Context, id string) (*teamdomain.Invitation, error) {
    var model InvitationModel
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get invitation: %w", err)
    }
    return model.ToDomain(), nil
}

func (r *TeamRepository) GetInvitationByToken(ctx context.Context, token string) (*teamdomain.Invitation, error) {
    var model InvitationModel
    if err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get invitation: %w", err)
    }
    return model.ToDomain(), nil
}

func (r *TeamRepository) GetInvitationByEmailAndTeam(ctx context.Context, email, teamID string) (*teamdomain.Invitation, error) {
    var model InvitationModel
    if err := r.db.WithContext(ctx).
        Where("email = ? AND team_id = ?", email, teamID).
        First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to get invitation: %w", err)
    }
    return model.ToDomain(), nil
}

func (r *TeamRepository) GetInvitationsByTeam(ctx context.Context, teamID string, filters teamdomain.ListInvitationsFilters) ([]*teamdomain.Invitation, int64, error) {
    query := r.db.WithContext(ctx).Model(&InvitationModel{}).Where("team_id = ?", teamID)

    // Apply filters
    if filters.Email != "" {
        query = query.Where("email = ?", filters.Email)
    }
    if filters.Status != "" {
        query = query.Where("status = ?", filters.Status)
    }

    // Count total
    var total int64
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count invitations: %w", err)
    }

    // Pagination
    if filters.Limit > 0 {
        query = query.Limit(filters.Limit)
        if filters.Offset > 0 {
            query = query.Offset(filters.Offset)
        }
    }

    var models []InvitationModel
    if err := query.Find(&models).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to get invitations: %w", err)
    }

    invitations := make([]*teamdomain.Invitation, len(models))
    for i, model := range models {
        invitations[i] = model.ToDomain()
    }
    return invitations, total, nil
}

func (r *TeamRepository) UpdateInvitation(ctx context.Context, invitation *teamdomain.Invitation) error {
    model := InvitationModel{}
    model.FromDomain(invitation)
    return r.db.WithContext(ctx).Save(&model).Error
}

func (r *TeamRepository) DeleteInvitation(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Model(&InvitationModel{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (r *TeamRepository) GetPendingInvitations(ctx context.Context) ([]*teamdomain.Invitation, error) {
    var models []InvitationModel
    if err := r.db.WithContext(ctx).
        Where("status = ? AND expires_at > NOW()", teamdomain.InvitationStatusPending).
        Find(&models).Error; err != nil {
        return nil, fmt.Errorf("failed to get pending invitations: %w", err)
    }

    invitations := make([]*teamdomain.Invitation, len(models))
    for i, model := range models {
        invitations[i] = model.ToDomain()
    }
    return invitations, nil
}