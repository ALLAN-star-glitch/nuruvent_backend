package postgres

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/google/uuid"
)

// ============================================================
// authdomain → DATABASE MODEL MAPPERS
// ============================================================

// ============================================================
// USER MAPPERS (formerly Account)
// ============================================================

func ToUserModel(user *authdomain.User) *UserModel {
	if user == nil {
		return nil
	}

	accountTypeID := uuid.MustParse(user.AccountTypeID)

	var professionalTypeID *uuid.UUID
	if user.ProfessionalTypeID != nil {
		id := uuid.MustParse(*user.ProfessionalTypeID)
		professionalTypeID = &id
	}

	var institutionID *uuid.UUID
	if user.InstitutionID != nil {
		id := uuid.MustParse(*user.InstitutionID)
		institutionID = &id
	}

	return &UserModel{
		ID:                 uuid.MustParse(user.ID),
		Slug:               user.Slug,
		Name:               user.Name,
		DisplayName:        user.DisplayName,
		Email:              user.Email,
		PasswordHash:       user.PasswordHash,
		Phone:              user.Phone,
		AccountTypeID:      accountTypeID,
		ProfessionalTypeID: professionalTypeID,
		InstitutionID:      institutionID,
		EmailVerified:      user.EmailVerified,
		EmailVerifiedAt:    user.EmailVerifiedAt,
		IdentityVerified:   user.IdentityVerified,
		PhoneVerified:      false,
		KYCStatus:          "pending",
		IsActive:           user.IsActive,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}
}

// ============================================================
// REFRESH TOKEN MAPPERS
// ============================================================

func ToRefreshTokenModel(token *authdomain.RefreshToken) *RefreshTokenModel {
	if token == nil {
		return nil
	}

	return &RefreshTokenModel{
		ID:        uuid.MustParse(token.ID),
		UserID:    uuid.MustParse(token.UserID), // formerly AccountID
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
		Revoked:   token.Revoked,
		UserAgent: token.UserAgent,
		IPAddress: token.IPAddress,
		CreatedAt: token.CreatedAt,
		UpdatedAt: token.UpdatedAt,
	}
}

// ============================================================
// INSTITUTION MAPPERS
// ============================================================

func ToInstitutionModel(institution *authdomain.Institution) *InstitutionModel {
	if institution == nil {
		return nil
	}

	return &InstitutionModel{
		ID:                uuid.MustParse(institution.ID),
		Slug:              institution.Slug,
		Name:              institution.Name,
		DisplayName:       institution.DisplayName,
		Email:             institution.Email,
		Phone:             institution.Phone,
		InstitutionTypeID: uuid.MustParse(institution.InstitutionTypeID),
		Description:       institution.Description,
		Logo:              institution.Logo,
		Website:           institution.Website,
		Address:           institution.Address,
		IsActive:          institution.IsActive,
		CreatedAt:         institution.CreatedAt,
		UpdatedAt:         institution.UpdatedAt,
	}
}

// ============================================================
// TEAM MEMBER MAPPERS (UPDATED)
// ============================================================

func ToTeamMemberModel(member *authdomain.TeamMember) *TeamMemberModel {
	if member == nil {
		return nil
	}

	model := &TeamMemberModel{
		ID:         uuid.MustParse(member.ID),
		MemberID:   uuid.MustParse(member.MemberID),
		TeamTypeID: uuid.MustParse(member.TeamTypeID),
		IsActive:   member.IsActive,
		JoinedAt:   member.JoinedAt,
		CreatedAt:  member.CreatedAt,
		UpdatedAt:  member.UpdatedAt,
	}

	// Handle nullable InstitutionID
	if member.InstitutionID != nil {
		id := uuid.MustParse(*member.InstitutionID)
		model.InstitutionID = &id
	}

	// Handle nullable InvitedBy
	if member.InvitedBy != nil {
		id := uuid.MustParse(*member.InvitedBy)
		model.InvitedBy = &id
	}

	// Handle nullable CreatedBy
	if member.CreatedBy != nil {
		id := uuid.MustParse(*member.CreatedBy)
		model.CreatedBy = &id
	}

	return model
}

// ============================================================
// TEAM TYPE MAPPERS (NEW)
// ============================================================

func ToTeamTypeModel(teamType *authdomain.TeamType) *TeamTypeModel {
	if teamType == nil {
		return nil
	}

	return &TeamTypeModel{
		ID:          uuid.MustParse(teamType.ID),
		Name:        teamType.Name,
		DisplayName: teamType.DisplayName,
		Slug:        teamType.Slug,
		Description: teamType.Description,
		IsActive:    teamType.IsActive,
		CreatedAt:   teamType.CreatedAt,
		UpdatedAt:   teamType.UpdatedAt,
	}
}

// ============================================================
// DATABASE MODEL → authdomain MAPPERS
// ============================================================

// ============================================================
// USER MAPPERS (formerly Account)
// ============================================================

func ToAuthDomainUser(model *UserModel) *authdomain.User {
	if model == nil {
		return nil
	}

	var professionalTypeID *string
	if model.ProfessionalTypeID != nil {
		id := model.ProfessionalTypeID.String()
		professionalTypeID = &id
	}

	var institutionID *string
	if model.InstitutionID != nil {
		id := model.InstitutionID.String()
		institutionID = &id
	}

	return &authdomain.User{
		ID:                 model.ID.String(),
		Slug:               model.Slug,
		Name:               model.Name,
		DisplayName:        model.DisplayName,
		Email:              model.Email,
		PasswordHash:       model.PasswordHash,
		Phone:              model.Phone,
		AccountTypeID:      model.AccountTypeID.String(),
		ProfessionalTypeID: professionalTypeID,
		InstitutionID:      institutionID,
		EmailVerified:      model.EmailVerified,
		EmailVerifiedAt:    model.EmailVerifiedAt,
		IdentityVerified:   model.IdentityVerified,
		IsActive:           model.IsActive,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

// ============================================================
// REFRESH TOKEN MAPPERS
// ============================================================

func ToAuthDomainRefreshToken(model *RefreshTokenModel) *authdomain.RefreshToken {
	if model == nil {
		return nil
	}

	return &authdomain.RefreshToken{
		ID:        model.ID.String(),
		UserID:    model.UserID.String(), // formerly AccountID
		Token:     model.Token,
		ExpiresAt: model.ExpiresAt,
		Revoked:   model.Revoked,
		UserAgent: model.UserAgent,
		IPAddress: model.IPAddress,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

// ============================================================
// INSTITUTION MAPPERS
// ============================================================

func ToAuthDomainInstitution(model *InstitutionModel) *authdomain.Institution {
	if model == nil {
		return nil
	}

	return &authdomain.Institution{
		ID:                model.ID.String(),
		Slug:              model.Slug,
		Name:              model.Name,
		DisplayName:       model.DisplayName,
		Email:             model.Email,
		Phone:             model.Phone,
		InstitutionTypeID: model.InstitutionTypeID.String(),
		Description:       model.Description,
		Logo:              model.Logo,
		Website:           model.Website,
		Address:           model.Address,
		IsActive:          model.IsActive,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}

// ============================================================
// TEAM MEMBER MAPPERS (UPDATED)
// ============================================================

func ToAuthDomainTeamMember(model *TeamMemberModel) *authdomain.TeamMember {
	if model == nil {
		return nil
	}

	var institutionID *string
	if model.InstitutionID != nil {
		id := model.InstitutionID.String()
		institutionID = &id
	}

	var invitedBy *string
	if model.InvitedBy != nil {
		id := model.InvitedBy.String()
		invitedBy = &id
	}

	var createdBy *string
	if model.CreatedBy != nil {
		id := model.CreatedBy.String()
		createdBy = &id
	}

	return &authdomain.TeamMember{
		ID:            model.ID.String(),
		MemberID:      model.MemberID.String(),
		InstitutionID: institutionID,
		TeamTypeID:    model.TeamTypeID.String(),
		InvitedBy:     invitedBy,
		IsActive:      model.IsActive,
		JoinedAt:      model.JoinedAt,
		CreatedBy:     createdBy,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}

// ============================================================
// TEAM TYPE MAPPERS (NEW)
// ============================================================

func ToAuthDomainTeamType(model *TeamTypeModel) *authdomain.TeamType {
	if model == nil {
		return nil
	}

	return &authdomain.TeamType{
		ID:          model.ID.String(),
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Slug:        model.Slug,
		Description: model.Description,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// ============================================================
// VALUE OBJECT MAPPERS
// ============================================================

func ToAuthDomainAccountType(model *AccountTypeModel) *authdomain.AccountType {
	if model == nil {
		return nil
	}

	return &authdomain.AccountType{
		ID:          model.ID.String(),
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		Icon:        model.Icon,
		Color:       model.Color,
		IsActive:    model.IsActive,
	}
}

func ToAuthDomainProfessionalType(model *ProfessionalTypeModel) *authdomain.ProfessionalType {
	if model == nil {
		return nil
	}

	return &authdomain.ProfessionalType{
		ID:          model.ID.String(),
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		IsActive:    model.IsActive,
	}
}

func ToAuthDomainInstitutionType(model *InstitutionTypeModel) *authdomain.InstitutionType {
	if model == nil {
		return nil
	}

	return &authdomain.InstitutionType{
		ID:          model.ID.String(),
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		IsActive:    model.IsActive,
	}
}

// ============================================================
// BULK MAPPERS
// ============================================================

func ToAuthDomainUsers(models []UserModel) []*authdomain.User {
	if len(models) == 0 {
		return nil
	}

	users := make([]*authdomain.User, len(models))
	for i, model := range models {
		users[i] = ToAuthDomainUser(&model)
	}
	return users
}

func ToAuthDomainTeamMembers(models []TeamMemberModel) []*authdomain.TeamMember {
	if len(models) == 0 {
		return nil
	}

	members := make([]*authdomain.TeamMember, len(models))
	for i, model := range models {
		members[i] = ToAuthDomainTeamMember(&model)
	}
	return members
}

func ToAuthDomainInstitutions(models []InstitutionModel) []*authdomain.Institution {
	if len(models) == 0 {
		return nil
	}

	institutions := make([]*authdomain.Institution, len(models))
	for i, model := range models {
		institutions[i] = ToAuthDomainInstitution(&model)
	}
	return institutions
}

func ToAuthDomainTeamTypes(models []TeamTypeModel) []*authdomain.TeamType {
	if len(models) == 0 {
		return nil
	}

	teamTypes := make([]*authdomain.TeamType, len(models))
	for i, model := range models {
		teamTypes[i] = ToAuthDomainTeamType(&model)
	}
	return teamTypes
}