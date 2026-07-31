package bizservice

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/models"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business/bizrepo"
	"github.com/google/uuid"
)

type BusinessService struct {
	businessRepo *bizrepo.BusinessRepository
	memberRepo   *bizrepo.BusinessMemberRepository
	enforcer     *authorization.Enforcer
	permService  *authorization.Service
}

func NewBusinessService(
	businessRepo *bizrepo.BusinessRepository,
	memberRepo *bizrepo.BusinessMemberRepository,
	enforcer *authorization.Enforcer,
	permService *authorization.Service,
) *BusinessService {
	return &BusinessService{
		businessRepo: businessRepo,
		memberRepo:   memberRepo,
		enforcer:     enforcer,
		permService:  permService,
	}
}

// ================================================
// BUSINESS CRUD
// ================================================

// CreateBusinessWithAdmin creates a business and assigns the creator as business_admin
func (s *BusinessService) CreateBusinessWithAdmin(ctx context.Context, userID uuid.UUID, business *models.Business) (*models.Business, error) {
	// Set created by
	business.CreatedBy = userID

	// Create business
	if err := s.businessRepo.Create(business); err != nil {
		return nil, fmt.Errorf("failed to create business: %w", err)
	}

	// Assign business_admin role to the creator
	if err := s.permService.AssignBusinessAdminRole(ctx, userID.String(), business.ID.String()); err != nil {
		return nil, fmt.Errorf("failed to assign business_admin role: %w", err)
	}

	// Create business member record
	member := &models.BusinessMember{
		BusinessID: business.ID,
		UserID:     userID,
		Role:       "business_admin",
		IsActive:   true,
	}
	if err := s.memberRepo.Create(member); err != nil {
		return nil, fmt.Errorf("failed to create business member: %w", err)
	}

	return business, nil
}

// CreateBusinessFromRegistration creates a business from registration data
func (s *BusinessService) CreateBusinessFromRegistration(ctx context.Context, userID uuid.UUID, data map[string]string) (*models.Business, error) {
	// Parse business type
	var businessTypeID uuid.UUID
	var err error

	if typeID, ok := data["business_type_id"]; ok && typeID != "" {
		businessTypeID, err = uuid.Parse(typeID)
		if err != nil {
			return nil, fmt.Errorf("invalid business_type_id: %w", err)
		}
	} else if typeName, ok := data["business_type"]; ok && typeName != "" {
		businessType, err := s.businessRepo.GetBusinessTypeByName(typeName)
		if err != nil {
			return nil, fmt.Errorf("failed to find business type: %w", err)
		}
		if businessType == nil {
			return nil, fmt.Errorf("invalid business type: %s", typeName)
		}
		businessTypeID = businessType.ID
	} else {
		return nil, fmt.Errorf("business type is required")
	}

	// Check business type category
	businessType, _ := s.businessRepo.GetBusinessTypeByID(businessTypeID)

	// Determine the type of business
	isOrganization := businessType != nil && businessType.Category == "organization"
	isFormalIndividual := businessType != nil && businessType.Name == "individual_formal"
	isInformalIndividual := businessType != nil && businessType.Name == "individual_informal"

	var business *models.Business

	if isOrganization {
		// Organization - use business data (business_email and business_phone)
		businessName := data["business_name"]
		if businessName == "" {
			businessName = data["name"]
		}
		
		businessEmail := data["business_email"]
		if businessEmail == "" {
			businessEmail = data["email"]
		}
		
		businessPhone := data["business_phone"]
		if businessPhone == "" {
			businessPhone = data["phone"]
		}
		
		business = &models.Business{
			Name:             businessName,
			BusinessTypeID:   businessTypeID,
			Email:            businessEmail,
			Phone:            businessPhone,
			Address:          data["business_address"],
			Description:      data["business_description"],
			IsActive:         true,
			IsEmailVerified:  data["business_email"] != "",
			IsVerified:       false,
			CreatedBy:        userID,
		}
	} else if isFormalIndividual {
		// Formal Individual - use business data (business_email and business_phone)
		businessName := data["business_name"]
		if businessName == "" {
			businessName = data["name"] // Fallback to personal name
		}

		businessEmail := data["business_email"]
		if businessEmail == "" {
			businessEmail = data["email"] // Fallback to personal email
		}

		businessPhone := data["business_phone"]
		if businessPhone == "" {
			businessPhone = data["phone"] // Fallback to personal phone
		}

		business = &models.Business{
			Name:           businessName,
			BusinessTypeID: businessTypeID,
			Email:          businessEmail,
			Phone:          businessPhone,
			Address:        data["business_address"],
			Description:    data["business_description"],
			IsActive:       true,
			IsEmailVerified: data["business_email"] != "",
			IsVerified:     false,
			CreatedBy:      userID,
		}
	} else if isInformalIndividual {
		// Informal Individual - use personal data (email and phone)
		// CRITICAL: Get values directly from data map
		name := data["name"]
		email := data["email"]
		phone := data["phone"]
		
		
		// Validate
		if name == "" {
			log.Printf("WARNING: name is empty for informal individual business, using fallback")
			name = "User"
		}
		
		if email == "" {
			log.Printf("ERROR: email is empty for informal individual business")
			return nil, fmt.Errorf("email is required for informal individual business")
		}
		
		if phone == "" {
			log.Printf("WARNING: phone is empty for informal individual business, using fallback")
			phone = "N/A"
		}

		business = &models.Business{
			Name:            name,
			BusinessTypeID:  businessTypeID,
			Email:           email,
			Phone:           phone,
			Address:         data["business_address"],
			Description:     data["business_description"],
			IsActive:        true,
			IsEmailVerified: true, // Already verified via OTP
			IsVerified:      false,
			CreatedBy:       userID,
		}
	} else {
		return nil, fmt.Errorf("unknown business type category")
	}

	return s.CreateBusinessWithAdmin(ctx, userID, business)
}

// GetBusinessByID gets a business by ID with permission check
func (s *BusinessService) GetBusinessByID(ctx context.Context, userID, businessID uuid.UUID) (*models.Business, error) {
	// Check permission
	canRead, err := s.enforcer.Enforce(userID.String(), authorization.BusinessDomain(businessID.String()), authorization.ResourceBusiness.String(), authorization.ActionRead.String())
	if err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	if !canRead {
		return nil, fmt.Errorf("insufficient permissions to view business")
	}

	return s.businessRepo.GetByID(businessID)
}

// GetBusinessByIDPublic gets a business by ID (public access, no permission check)
func (s *BusinessService) GetBusinessByIDPublic(ctx context.Context, businessID uuid.UUID) (*models.Business, error) {
	return s.businessRepo.GetByID(businessID)
}

// GetBusinessBySlug gets a business by slug (public)
func (s *BusinessService) GetBusinessBySlug(ctx context.Context, slug string) (*models.Business, error) {
	return s.businessRepo.GetBySlug(slug)
}

// GetBusinessByEmail gets a business by email
func (s *BusinessService) GetBusinessByEmail(ctx context.Context, email string) (*models.Business, error) {
	return s.businessRepo.GetByEmail(email)
}

// UpdateBusiness updates a business with permission check
func (s *BusinessService) UpdateBusiness(ctx context.Context, userID, businessID uuid.UUID, updates map[string]interface{}) (*models.Business, error) {
	// Check permission
	canUpdate, err := s.enforcer.Enforce(userID.String(), authorization.BusinessDomain(businessID.String()), authorization.ResourceBusiness.String(), authorization.ActionUpdate.String())
	if err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	if !canUpdate {
		return nil, fmt.Errorf("insufficient permissions to update business")
	}

	business, err := s.businessRepo.GetByID(businessID)
	if err != nil {
		return nil, err
	}
	if business == nil {
		return nil, fmt.Errorf("business not found")
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok && name != "" {
		business.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		business.Description = description
	}
	if email, ok := updates["email"].(string); ok {
		business.Email = email
	}
	if phone, ok := updates["phone"].(string); ok {
		business.Phone = phone
	}
	if address, ok := updates["address"].(string); ok {
		business.Address = address
	}
	if logo, ok := updates["logo"].(string); ok {
		business.Logo = logo
	}
	if website, ok := updates["website"].(string); ok {
		business.Website = website
	}
	if businessTypeID, ok := updates["business_type_id"].(string); ok {
		id, err := uuid.Parse(businessTypeID)
		if err == nil {
			business.BusinessTypeID = id
		}
	}

	if err := s.businessRepo.Update(business); err != nil {
		return nil, fmt.Errorf("failed to update business: %w", err)
	}

	return business, nil
}

// DeleteBusiness deletes a business with permission check
func (s *BusinessService) DeleteBusiness(ctx context.Context, userID, businessID uuid.UUID) error {
	// Check permission
	canDelete, err := s.enforcer.Enforce(userID.String(), authorization.BusinessDomain(businessID.String()), authorization.ResourceBusiness.String(), authorization.ActionDelete.String())
	if err != nil {
		return fmt.Errorf("authorization error: %w", err)
	}
	if !canDelete {
		return fmt.Errorf("insufficient permissions to delete business")
	}

	// Remove all business policies
	if err := s.permService.RemoveBusinessPolicies(ctx, businessID.String()); err != nil {
		return fmt.Errorf("failed to remove business policies: %w", err)
	}

	// Delete business
	if err := s.businessRepo.Delete(businessID); err != nil {
		return fmt.Errorf("failed to delete business: %w", err)
	}

	return nil
}

// ================================================
// QUERY OPERATIONS
// ================================================

// GetMyBusinesses gets all businesses a user belongs to
func (s *BusinessService) GetMyBusinesses(ctx context.Context, userID uuid.UUID) ([]models.Business, error) {
	return s.businessRepo.GetBusinessesByUser(userID)
}

// SearchBusinesses searches businesses
func (s *BusinessService) SearchBusinesses(ctx context.Context, query string, page, pageSize int) ([]models.Business, int64, error) {
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.businessRepo.Search(query, limit, offset)
}

// GetAllBusinesses gets all businesses with pagination
func (s *BusinessService) GetAllBusinesses(ctx context.Context, page, pageSize int) ([]models.Business, int64, error) {
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.businessRepo.GetAllBusinesses(limit, offset)
}

// GetBusinessStats gets business statistics with permission check
func (s *BusinessService) GetBusinessStats(ctx context.Context, userID, businessID uuid.UUID) (map[string]interface{}, error) {
	// Check read permission
	canRead, err := s.enforcer.Enforce(userID.String(), authorization.BusinessDomain(businessID.String()), authorization.ResourceBusiness.String(), authorization.ActionRead.String())
	if err != nil {
		return nil, fmt.Errorf("authorization error: %w", err)
	}
	if !canRead {
		return nil, fmt.Errorf("insufficient permissions to view business stats")
	}

	return s.businessRepo.GetBusinessStats(businessID)
}

// ================================================
// BUSINESS TYPES
// ================================================

// GetBusinessTypes gets all business types
func (s *BusinessService) GetBusinessTypes(ctx context.Context) ([]models.BusinessType, error) {
	return s.businessRepo.GetAllBusinessTypes()
}

// GetBusinessTypeByName gets a business type by name
func (s *BusinessService) GetBusinessTypeByName(ctx context.Context, name string) (*models.BusinessType, error) {
	return s.businessRepo.GetBusinessTypeByName(name)
}

// GetBusinessTypeByID gets a business type by ID
func (s *BusinessService) GetBusinessTypeByID(ctx context.Context, id uuid.UUID) (*models.BusinessType, error) {
	return s.businessRepo.GetBusinessTypeByID(id)
}

// ================================================
// BUSINESS VERIFICATION
// ================================================

// VerifyBusinessEmail marks a business email as verified
func (s *BusinessService) VerifyBusinessEmail(ctx context.Context, businessID uuid.UUID) error {
	business, err := s.businessRepo.GetByID(businessID)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("business not found")
	}

	business.IsEmailVerified = true
	now := time.Now()
	business.EmailVerifiedAt = &now

	return s.businessRepo.Update(business)
}

// VerifyBusiness marks a business as verified (by admin)
func (s *BusinessService) VerifyBusiness(ctx context.Context, adminID, businessID uuid.UUID) error {
	// Check if admin has permission
	canManage, err := s.enforcer.Enforce(adminID.String(), authorization.DomainPlatform, authorization.ResourceBusiness.String(), authorization.ActionUpdate.String())
	if err != nil {
		return fmt.Errorf("authorization error: %w", err)
	}
	if !canManage {
		return fmt.Errorf("insufficient permissions to verify business")
	}

	business, err := s.businessRepo.GetByID(businessID)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("business not found")
	}

	business.IsVerified = true
	now := time.Now()
	business.VerifiedAt = &now
	business.VerifiedBy = &adminID

	return s.businessRepo.Update(business)
}