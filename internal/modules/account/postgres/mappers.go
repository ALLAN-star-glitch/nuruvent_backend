// internal/modules/account/postgres/mappers.go

package postgres

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/account/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"
)

// ============================================================
// ACCOUNT MAPPERS
// ============================================================

// ToDomainAccount converts AccountModel to domain.Account
func ToDomainAccount(model *AccountModel) *domain.Account {
	if model == nil {
		return nil
	}
	return &domain.Account{
		ID:                 model.ID,
		Slug:               model.Slug,
		Name:               model.Name,
		DisplayName:        model.DisplayName,
		Email:              model.Email,
		PasswordHash:       model.PasswordHash,
		Phone:              model.Phone,
		AccountTypeID:      model.AccountTypeID,
		ProfessionalTypeID: model.ProfessionalTypeID,
		InstitutionID:      model.InstitutionID,
		EmailVerified:      model.EmailVerified,
		EmailVerifiedAt:    model.EmailVerifiedAt,
		PhoneVerified:      model.PhoneVerified,
		PhoneVerifiedAt:    model.PhoneVerifiedAt,
		IdentityVerified:   model.IdentityVerified,
		KYCStatus:          model.KYCStatus,
		IsActive:           model.IsActive,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

// ToModelAccount converts domain.Account to AccountModel
func ToModelAccount(account *domain.Account) *AccountModel {
	if account == nil {
		return nil
	}
	return &AccountModel{
		ID:                 account.ID,
		Slug:               account.Slug,
		Name:               account.Name,
		DisplayName:        account.DisplayName,
		Email:              account.Email,
		PasswordHash:       account.PasswordHash,
		Phone:              account.Phone,
		AccountTypeID:      account.AccountTypeID,
		ProfessionalTypeID: account.ProfessionalTypeID,
		InstitutionID:      account.InstitutionID,
		EmailVerified:      account.EmailVerified,
		EmailVerifiedAt:    account.EmailVerifiedAt,
		PhoneVerified:      account.PhoneVerified,
		PhoneVerifiedAt:    account.PhoneVerifiedAt,
		IdentityVerified:   account.IdentityVerified,
		KYCStatus:          account.KYCStatus,
		IsActive:           account.IsActive,
		CreatedAt:          account.CreatedAt,
		UpdatedAt:          account.UpdatedAt,
	}
}

// ============================================================
// ACCOUNT TYPE INFO MAPPERS (Value Object)
// ============================================================

// ToDomainAccountTypeInfo converts AccountTypeModel to domain.AccountTypeInfo
func ToDomainAccountTypeInfo(model *AccountTypeModel) (*domain.AccountTypeInfo, error) {
	if model == nil {
		return nil, nil
	}

	// ✅ Use shared types to parse
	accountType, valid := types.ParseAccountType(model.Name)
	if !valid {
		// Try parsing by slug
		accountType, valid = types.ParseAccountTypeBySlug(model.Slug)
		if !valid {
			return nil, domain.ErrInvalidAccountType
		}
	}

	// ✅ Use domain helper to get info
	info, ok := domain.GetAccountTypeInfo(accountType)
	if !ok {
		return nil, domain.ErrAccountTypeNotFound
	}

	// Override with database values
	info.Name = model.Name
	info.DisplayName = model.DisplayName
	info.Description = model.Description
	info.Icon = model.Icon
	info.Color = model.Color
	info.SortOrder = model.SortOrder
	info.IsActive = model.IsActive

	return &info, nil
}

// ToModelAccountType converts domain.AccountTypeInfo to AccountTypeModel
func ToModelAccountType(info domain.AccountTypeInfo) *AccountTypeModel {
	return &AccountTypeModel{
		Slug:        info.Slug, // ✅ info.Slug is already a string
		Name:        info.Name,
		DisplayName: info.DisplayName,
		Description: info.Description,
		Icon:        info.Icon,
		Color:       info.Color,
		SortOrder:   info.SortOrder,
		IsActive:    info.IsActive,
	}
}

// ============================================================
// ACCOUNT TYPE ENTITY MAPPERS (Domain Entity)
// ============================================================

// ToDomainAccountTypeEntity converts AccountTypeModel to domain.AccountType
func ToDomainAccountTypeEntity(model *AccountTypeModel) *domain.AccountType {
	if model == nil {
		return nil
	}
	return domain.NewAccountTypeFromRepository(
		model.ID,
		model.Slug,
		model.Name,
		model.DisplayName,
		model.Description,
		model.Icon,
		model.Color,
		model.SortOrder,
		model.IsActive,
		model.CreatedAt,
		model.UpdatedAt,
		nil,
	)
}

// ToModelAccountTypeEntity converts domain.AccountType to AccountTypeModel
func ToModelAccountTypeEntity(accountType *domain.AccountType) *AccountTypeModel {
	if accountType == nil {
		return nil
	}
	return &AccountTypeModel{
		ID:          accountType.ID,
		Slug:        accountType.Slug,
		Name:        accountType.Name,
		DisplayName: accountType.DisplayName,
		Description: accountType.Description,
		Icon:        accountType.Icon,
		Color:       accountType.Color,
		SortOrder:   accountType.SortOrder,
		IsActive:    accountType.IsActive,
	}
}

// ============================================================
// PROFESSIONAL TYPE INFO MAPPERS (Value Object)
// ============================================================

// ToDomainProfessionalTypeInfo converts ProfessionalTypeModel to domain.ProfessionalTypeInfo
func ToDomainProfessionalTypeInfo(model *ProfessionalTypeModel) (*domain.ProfessionalTypeInfo, error) {
	if model == nil {
		return nil, nil
	}

	// ✅ Use shared types to parse
	profType, valid := types.ParseProfessionalType(model.Name)
	if !valid {
		// Try parsing by slug
		profType, valid = types.ParseProfessionalTypeBySlug(model.Slug)
		if !valid {
			return nil, domain.ErrInvalidProfessionalType
		}
	}

	// ✅ Use domain helper to get info
	info, ok := domain.GetProfessionalTypeInfo(profType)
	if !ok {
		return nil, domain.ErrProfessionalTypeNotFound
	}

	info.Name = model.Name
	info.DisplayName = model.DisplayName
	info.Description = model.Description
	info.Icon = model.Icon
	info.Color = model.Color
	info.SortOrder = model.SortOrder
	info.CanHost = model.CanHost
	info.IsActive = model.IsActive

	return &info, nil
}

// ToModelProfessionalType converts domain.ProfessionalTypeInfo to ProfessionalTypeModel
func ToModelProfessionalType(info domain.ProfessionalTypeInfo) *ProfessionalTypeModel {
	return &ProfessionalTypeModel{
		Slug:        info.Slug, // ✅ info.Slug is already a string
		Name:        info.Name,
		DisplayName: info.DisplayName,
		Description: info.Description,
		Icon:        info.Icon,
		Color:       info.Color,
		SortOrder:   info.SortOrder,
		CanHost:     info.CanHost,
		IsActive:    info.IsActive,
	}
}

// ============================================================
// PROFESSIONAL TYPE ENTITY MAPPERS (Domain Entity)
// ============================================================

// ToDomainProfessionalTypeEntity converts ProfessionalTypeModel to domain.ProfessionalType
func ToDomainProfessionalTypeEntity(model *ProfessionalTypeModel) *domain.ProfessionalType {
	if model == nil {
		return nil
	}
	return domain.NewProfessionalTypeFromRepository(
		model.ID,
		model.Slug,
		model.Name,
		model.DisplayName,
		model.Description,
		model.Icon,
		model.Color,
		model.SortOrder,
		model.CanHost,
		model.IsActive,
		model.CreatedAt,
		model.UpdatedAt,
		nil,
	)
}

// ToModelProfessionalTypeEntity converts domain.ProfessionalType to ProfessionalTypeModel
func ToModelProfessionalTypeEntity(profType *domain.ProfessionalType) *ProfessionalTypeModel {
	if profType == nil {
		return nil
	}
	return &ProfessionalTypeModel{
		ID:          profType.ID,
		Slug:        profType.Slug,
		Name:        profType.Name,
		DisplayName: profType.DisplayName,
		Description: profType.Description,
		Icon:        profType.Icon,
		Color:       profType.Color,
		SortOrder:   profType.SortOrder,
		CanHost:     profType.CanHost,
		IsActive:    profType.IsActive,
	}
}

// ============================================================
// INSTITUTION TYPE INFO MAPPERS (Value Object)
// ============================================================

// ToDomainInstitutionTypeInfo converts InstitutionTypeModel to domain.InstitutionTypeInfo
func ToDomainInstitutionTypeInfo(model *InstitutionTypeModel) (*domain.InstitutionTypeInfo, error) {
	if model == nil {
		return nil, nil
	}

	// ✅ Use shared types to parse
	instType, valid := types.ParseInstitutionType(model.Name)
	if !valid {
		// Try parsing by slug
		instType, valid = types.ParseInstitutionTypeBySlug(model.Slug)
		if !valid {
			return nil, domain.ErrInvalidInstitutionType
		}
	}

	// ✅ Use domain helper to get info
	info, ok := domain.GetInstitutionTypeInfo(instType)
	if !ok {
		return nil, domain.ErrInstitutionTypeNotFound
	}

	info.Name = model.Name
	info.DisplayName = model.DisplayName
	info.Description = model.Description
	info.Icon = model.Icon
	info.Color = model.Color
	info.SortOrder = model.SortOrder
	info.IsActive = model.IsActive

	return &info, nil
}

// ToModelInstitutionType converts domain.InstitutionTypeInfo to InstitutionTypeModel
func ToModelInstitutionType(info domain.InstitutionTypeInfo) *InstitutionTypeModel {
	return &InstitutionTypeModel{
		Slug:        info.Slug, // ✅ info.Slug is already a string
		Name:        info.Name,
		DisplayName: info.DisplayName,
		Description: info.Description,
		Icon:        info.Icon,
		Color:       info.Color,
		SortOrder:   info.SortOrder,
		IsActive:    info.IsActive,
	}
}

// ============================================================
// INSTITUTION TYPE ENTITY MAPPERS (Domain Entity)
// ============================================================

// ToDomainInstitutionTypeEntity converts InstitutionTypeModel to domain.InstitutionType
func ToDomainInstitutionTypeEntity(model *InstitutionTypeModel) *domain.InstitutionType {
	if model == nil {
		return nil
	}
	return domain.NewInstitutionTypeFromRepository(
		model.ID,
		model.Slug,
		model.Name,
		model.DisplayName,
		model.Description,
		model.Icon,
		model.Color,
		model.SortOrder,
		model.IsActive,
		model.CreatedAt,
		model.UpdatedAt,
		nil,
	)
}

// ToModelInstitutionTypeEntity converts domain.InstitutionType to InstitutionTypeModel
func ToModelInstitutionTypeEntity(instType *domain.InstitutionType) *InstitutionTypeModel {
	if instType == nil {
		return nil
	}
	return &InstitutionTypeModel{
		ID:          instType.ID,
		Slug:        instType.Slug,
		Name:        instType.Name,
		DisplayName: instType.DisplayName,
		Description: instType.Description,
		Icon:        instType.Icon,
		Color:       instType.Color,
		SortOrder:   instType.SortOrder,
		IsActive:    instType.IsActive,
	}
}

// ============================================================
// INSTITUTION MAPPERS
// ============================================================

// ToDomainInstitution converts InstitutionModel to domain.Institution
func ToDomainInstitution(model *InstitutionModel) *domain.Institution {
	if model == nil {
		return nil
	}
	return &domain.Institution{
		ID:                model.ID,
		Slug:              model.Slug,
		Name:              model.Name,
		DisplayName:       model.DisplayName,
		Email:             model.Email,
		Phone:             model.Phone,
		InstitutionTypeID: model.InstitutionTypeID,
		Description:       model.Description,
		Logo:              model.Logo,
		Website:           model.Website,
		Address:           model.Address,
		IsActive:          model.IsActive,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}
}

// ToModelInstitution converts domain.Institution to InstitutionModel
func ToModelInstitution(institution *domain.Institution) *InstitutionModel {
	if institution == nil {
		return nil
	}
	return &InstitutionModel{
		ID:                institution.ID,
		Slug:              institution.Slug,
		Name:              institution.Name,
		DisplayName:       institution.DisplayName,
		Email:             institution.Email,
		Phone:             institution.Phone,
		InstitutionTypeID: institution.InstitutionTypeID,
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
// TEAM MEMBER MAPPERS
// ============================================================

// ToDomainTeamMember converts TeamMemberModel to domain.TeamMember
func ToDomainTeamMember(model *TeamMemberModel) *domain.TeamMember {
	if model == nil {
		return nil
	}
	return &domain.TeamMember{
		ID:          model.ID,
		Slug:        model.Slug,
		Name:        model.Name,
		DisplayName: model.DisplayName,
		AccountID:   model.AccountID,
		MemberID:    model.MemberID,
		Role:        model.Role,
		JobTitle:    model.JobTitle,
		IsActive:    model.IsActive,
		JoinedAt:    model.JoinedAt,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

// ToModelTeamMember converts domain.TeamMember to TeamMemberModel
func ToModelTeamMember(member *domain.TeamMember) *TeamMemberModel {
	if member == nil {
		return nil
	}
	return &TeamMemberModel{
		ID:          member.ID,
		Slug:        member.Slug,
		Name:        member.Name,
		DisplayName: member.DisplayName,
		AccountID:   member.AccountID,
		MemberID:    member.MemberID,
		Role:        member.Role,
		JobTitle:    member.JobTitle,
		IsActive:    member.IsActive,
		JoinedAt:    member.JoinedAt,
		CreatedAt:   member.CreatedAt,
		UpdatedAt:   member.UpdatedAt,
	}
}