package postgres

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
	"github.com/google/uuid"
)

// ============================================================
// authdomain → DATABASE MODEL MAPPERS
// ============================================================

func ToAccountModel(account *authdomain.Account) *AccountModel {
	if account == nil {
		return nil
	}

	accountTypeID := uuid.MustParse(account.AccountTypeID)

	var professionalTypeID *uuid.UUID
	if account.ProfessionalTypeID != nil {
		id := uuid.MustParse(*account.ProfessionalTypeID)
		professionalTypeID = &id
	}

	var institutionID *uuid.UUID
	if account.InstitutionID != nil {
		id := uuid.MustParse(*account.InstitutionID)
		institutionID = &id
	}

	return &AccountModel{
		ID:                 uuid.MustParse(account.ID),
		Slug:               account.Slug,
		Name:               account.Name,
		DisplayName:        account.DisplayName,
		Email:              account.Email,
		PasswordHash:       account.PasswordHash,
		Phone:              account.Phone,
		AccountTypeID:      accountTypeID,
		ProfessionalTypeID: professionalTypeID,
		InstitutionID:      institutionID,
		EmailVerified:      account.EmailVerified,
		EmailVerifiedAt:    account.EmailVerifiedAt,
		IdentityVerified:   account.IdentityVerified,
		IsActive:           account.IsActive,
		CreatedAt:          account.CreatedAt,
		UpdatedAt:          account.UpdatedAt,
	}
}

func ToRefreshTokenModel(token *authdomain.RefreshToken) *RefreshTokenModel {
	if token == nil {
		return nil
	}

	return &RefreshTokenModel{
		ID:         uuid.MustParse(token.ID),
		AccountID:  uuid.MustParse(token.AccountID),
		Token:      token.Token,
		ExpiresAt:  token.ExpiresAt,
		Revoked:    token.Revoked,
		UserAgent:  token.UserAgent,
		IPAddress:  token.IPAddress,
		CreatedAt:  token.CreatedAt,
		UpdatedAt:  token.UpdatedAt,
	}
}

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

func ToTeamMemberModel(member *authdomain.TeamMember) *TeamMemberModel {
	if member == nil {
		return nil
	}

	model := &TeamMemberModel{
		ID:          uuid.MustParse(member.ID),
		Slug:        member.Slug,
		Name:        member.Name,
		DisplayName: member.DisplayName,
		AccountID:   uuid.MustParse(member.AccountID),
		MemberID:    uuid.MustParse(member.MemberID),
		Role:        TeamMemberRole(member.Role),
		JobTitle:    member.JobTitle,
		IsActive:    member.IsActive,
		JoinedAt:    member.JoinedAt,
		CreatedAt:   member.CreatedAt,
		UpdatedAt:   member.UpdatedAt,
	}

	// ✅ Handle nullable CreatedBy
	if member.CreatedBy != nil && *member.CreatedBy != "" {
		id := uuid.MustParse(*member.CreatedBy)
		model.CreatedBy = &id
	}

	return model
}

// ============================================================
// DATABASE MODEL → authdomain MAPPERS
// ============================================================

func ToauthdomainAccount(model *AccountModel) *authdomain.Account {
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

	return &authdomain.Account{
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

func ToauthdomainRefreshToken(model *RefreshTokenModel) *authdomain.RefreshToken {
	if model == nil {
		return nil
	}

	return &authdomain.RefreshToken{
		ID:         model.ID.String(),
		AccountID:  model.AccountID.String(),
		Token:      model.Token,
		ExpiresAt:  model.ExpiresAt,
		Revoked:    model.Revoked,
		UserAgent:  model.UserAgent,
		IPAddress:  model.IPAddress,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}

func ToauthdomainInstitution(model *InstitutionModel) *authdomain.Institution {
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

func ToauthdomainTeamMember(model *TeamMemberModel) *authdomain.TeamMember {
	if model == nil {
		return nil
	}

	var createdBy *string
	if model.CreatedBy != nil {
		id := model.CreatedBy.String()
		createdBy = &id
	}

	return &authdomain.TeamMember{
		ID:          model.ID.String(),
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		AccountID:   model.AccountID.String(),
		MemberID:    model.MemberID.String(),
		Role:        string(model.Role),
		JobTitle:    model.JobTitle,
		IsActive:    model.IsActive,
		CreatedBy:   createdBy,
		JoinedAt:    model.JoinedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// ============================================================
// VALUE OBJECT MAPPERS
// ============================================================

func ToauthdomainAccountType(model *AccountTypeModel) *authdomain.AccountType {
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

func ToauthdomainProfessionalType(model *ProfessionalTypeModel) *authdomain.ProfessionalType {
	if model == nil {
		return nil
	}

	return &authdomain.ProfessionalType{
		ID:          model.ID.String(),
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		CanHost:     model.CanHost,
		IsActive:    model.IsActive,
	}
}

func ToauthdomainInstitutionType(model *InstitutionTypeModel) *authdomain.InstitutionType {
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