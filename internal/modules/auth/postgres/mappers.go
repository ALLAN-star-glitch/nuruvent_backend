// internal/modules/auth/infrastructure/postgres/mappers.go

package postgres

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/authdomain"
)

// ============================================================
// authdomain → DATABASE MODEL MAPPERS
// ============================================================

// ============================================================
// USER MAPPERS
// ============================================================

func ToUserModel(user *authdomain.User) *UserModel {
	if user == nil {
		return nil
	}

	var professionalTypeID *string
	if user.ProfessionalTypeID != nil {
		professionalTypeID = user.ProfessionalTypeID
	}

	return &UserModel{
		ID:                 user.ID,
		Slug:               user.Slug,
		Name:               user.Name,
		DisplayName:        user.DisplayName,
		Email:              user.Email,
		PasswordHash:       user.PasswordHash,
		Phone:              user.Phone,
		AccountTypeID:      user.AccountTypeID,
		ProfessionalTypeID: professionalTypeID,
		EmailVerified:      user.EmailVerified,
		EmailVerifiedAt:    user.EmailVerifiedAt,
		IdentityVerified:   user.IdentityVerified,
		IdentityVerifiedAt: user.IdentityVerifiedAt,
		PhoneVerified:      user.PhoneVerified,
		PhoneVerifiedAt:    user.PhoneVerifiedAt,
		IsActive:           user.IsActive,
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}
}

func ToAuthDomainUser(model *UserModel) *authdomain.User {
	if model == nil {
		return nil
	}

	var professionalTypeID *string
	if model.ProfessionalTypeID != nil {
		professionalTypeID = model.ProfessionalTypeID
	}

	return &authdomain.User{
		ID:                 model.ID,
		Slug:               model.Slug,
		Name:               model.Name,
		DisplayName:        model.DisplayName,
		Email:              model.Email,
		PasswordHash:       model.PasswordHash,
		Phone:              model.Phone,
		AccountTypeID:      model.AccountTypeID,
		ProfessionalTypeID: professionalTypeID,
		EmailVerified:      model.EmailVerified,
		EmailVerifiedAt:    model.EmailVerifiedAt,
		IdentityVerified:   model.IdentityVerified,
		IdentityVerifiedAt: model.IdentityVerifiedAt,
		PhoneVerified:      model.PhoneVerified,
		PhoneVerifiedAt:    model.PhoneVerifiedAt,
		IsActive:           model.IsActive,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

// ============================================================
// ACCOUNT MAPPERS
// ============================================================

func ToAccountModel(account *authdomain.Account) *AccountModel {
	if account == nil {
		return nil
	}

	var institutionTypeID *string
	if account.InstitutionTypeID != nil && *account.InstitutionTypeID != "" {
		institutionTypeID = account.InstitutionTypeID
	}

	return &AccountModel{
		ID:                account.ID,
		Name:              account.Name,
		DisplayName:       account.DisplayName,
		Slug:              account.Slug,
		Email:             account.Email,
		Phone:             account.Phone,
		AccountTypeID:     account.AccountTypeID,
		InstitutionTypeID: institutionTypeID,
		Status:            account.Status,
		LogoURL:           account.LogoURL,
		Website:           account.Website,
		Description:       account.Description,
		Address:           account.Address,
		City:              account.City,
		Country:           account.Country,
		BillingEmail:      account.BillingEmail,
		SubscriptionPlan:  account.SubscriptionPlan,
		CreatedBy:         account.CreatedBy,
		CreatedAt:         account.CreatedAt,
		UpdatedAt:         account.UpdatedAt,
	}
}

func ToAuthDomainAccount(model *AccountModel) *authdomain.Account {
	if model == nil {
		return nil
	}

	var institutionTypeID *string
	if model.InstitutionTypeID != nil && *model.InstitutionTypeID != "" {
		institutionTypeID = model.InstitutionTypeID
	}

	return &authdomain.Account{
		ID:                model.ID,
		Name:              model.Name,
		DisplayName:       model.DisplayName,
		Slug:              model.Slug,
		Email:             model.Email,
		Phone:             model.Phone,
		AccountTypeID:     model.AccountTypeID,
		InstitutionTypeID: institutionTypeID,
		Status:            model.Status,
		LogoURL:           model.LogoURL,
		Website:           model.Website,
		Description:       model.Description,
		Address:           model.Address,
		City:              model.City,
		Country:           model.Country,
		BillingEmail:      model.BillingEmail,
		SubscriptionPlan:  model.SubscriptionPlan,
		CreatedBy:         model.CreatedBy,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
		DeletedAt:         model.DeletedAt,
	}
}

// ============================================================
// ACCOUNT MEMBER MAPPERS
// ============================================================

func ToAccountMemberModel(member *authdomain.AccountMember) *AccountMemberModel {
	if member == nil {
		return nil
	}

	return &AccountMemberModel{
		ID:          member.ID,
		AccountID:   member.AccountID,
		UserID:      member.UserID,
		Role:        member.Role,
		IsActive:    member.IsActive,
		InvitedBy:   member.InvitedBy,
		JoinedAt:    member.JoinedAt,
		CreatedAt:   member.CreatedAt,
		UpdatedAt:   member.UpdatedAt,
	}
}

func ToAuthDomainAccountMember(model *AccountMemberModel) *authdomain.AccountMember {
	if model == nil {
		return nil
	}

	return &authdomain.AccountMember{
		ID:          model.ID,
		AccountID:   model.AccountID,
		UserID:      model.UserID,
		Role:        model.Role,
		IsActive:    model.IsActive,
		InvitedBy:   model.InvitedBy,
		JoinedAt:    model.JoinedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		DeletedAt:   model.DeletedAt,
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
		ID:         token.ID,
		UserID:     token.UserID,
		Token:      token.Token,
		ExpiresAt:  token.ExpiresAt,
		Revoked:    token.Revoked,
		UserAgent:  token.UserAgent,
		IPAddress:  token.IPAddress,
		CreatedAt:  token.CreatedAt,
		UpdatedAt:  token.UpdatedAt,
		DeletedAt:  token.DeletedAt,
	}
}

func ToAuthDomainRefreshToken(model *RefreshTokenModel) *authdomain.RefreshToken {
	if model == nil {
		return nil
	}

	return &authdomain.RefreshToken{
		ID:         model.ID,
		UserID:     model.UserID,
		Token:      model.Token,
		ExpiresAt:  model.ExpiresAt,
		Revoked:    model.Revoked,
		UserAgent:  model.UserAgent,
		IPAddress:  model.IPAddress,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		DeletedAt:  model.DeletedAt,
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
		ID:          model.ID,
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		Icon:        model.Icon,
		Color:       model.Color,
		SortOrder:   model.SortOrder,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		DeletedAt:   &model.DeletedAt.Time,
	}
}

func ToAuthDomainProfessionalType(model *ProfessionalTypeModel) *authdomain.ProfessionalType {
	if model == nil {
		return nil
	}

	return &authdomain.ProfessionalType{
		ID:          model.ID,
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		Icon:        model.Icon,
		Color:       model.Color,
		SortOrder:   model.SortOrder,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		DeletedAt:   &model.DeletedAt.Time,
	}
}

func ToAuthDomainInstitutionType(model *InstitutionTypeModel) *authdomain.InstitutionType {
	if model == nil {
		return nil
	}

	return &authdomain.InstitutionType{
		ID:          model.ID,
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		Description: model.Description,
		Icon:        model.Icon,
		Color:       model.Color,
		SortOrder:   model.SortOrder,
		IsActive:    model.IsActive,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		DeletedAt:   &model.DeletedAt.Time,
	}
}

// ============================================================
// BULK MAPPERS
// ============================================================

func ToAuthDomainUsers(models []UserModel) []*authdomain.User {
	if len(models) == 0 {
		return []*authdomain.User{}
	}

	users := make([]*authdomain.User, len(models))
	for i := range models {
		users[i] = ToAuthDomainUser(&models[i])
	}
	return users
}

func ToAuthDomainAccounts(models []AccountModel) []*authdomain.Account {
	if len(models) == 0 {
		return []*authdomain.Account{}
	}

	accounts := make([]*authdomain.Account, len(models))
	for i := range models {
		accounts[i] = ToAuthDomainAccount(&models[i])
	}
	return accounts
}

func ToAuthDomainAccountMembers(models []AccountMemberModel) []*authdomain.AccountMember {
	if len(models) == 0 {
		return []*authdomain.AccountMember{}
	}

	members := make([]*authdomain.AccountMember, len(models))
	for i := range models {
		members[i] = ToAuthDomainAccountMember(&models[i])
	}
	return members
}