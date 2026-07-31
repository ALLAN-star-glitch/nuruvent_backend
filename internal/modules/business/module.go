package business

import (
	"context"
	"log"

	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/config"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/database"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/authorization"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business/bizhandler"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business/bizrepo"
	"github.com/ALLAN_star_glitch/nuruvent-backend/internal/modules/business/bizservice"
	"github.com/gofiber/fiber/v3"
)

type Module struct {
	businessHandler *bizhandler.BusinessHandler
	memberHandler   *bizhandler.MemberHandler
	
	businessService *bizservice.BusinessService
	memberService   *bizservice.MemberService
}

func NewModule(
	cfg *config.Config,
	permService *authorization.Service,
	enforcer *authorization.Enforcer,
) *Module {
	db := database.GetDB()

	// Initialize repositories
	businessRepo := bizrepo.NewBusinessRepository(db)
	memberRepo := bizrepo.NewBusinessMemberRepository(db)

	// Initialize services
	businessService := bizservice.NewBusinessService(
		businessRepo,
		memberRepo,
		enforcer,
		permService,
	)
	
	memberService := bizservice.NewMemberService(
		memberRepo,
		businessRepo,
		enforcer,
		permService,
	)

	// Initialize handlers
	businessHandler := bizhandler.NewBusinessHandler(businessService)
	memberHandler := bizhandler.NewMemberHandler(memberService)

	return &Module{
		businessHandler: businessHandler,
		memberHandler:   memberHandler,
		businessService: businessService,
		memberService:   memberService,
	}
}

// SetupRoutes registers all business routes
func (m *Module) SetupRoutes(
	router fiber.Router,
	enforcer *authorization.Enforcer,
) {
	RegisterBusinessRoutes(
		router,
		m.businessHandler,
		m.memberHandler,
		enforcer,
	)
}

// GetBusinessService returns the business service
func (m *Module) GetBusinessService() *bizservice.BusinessService {
	return m.businessService
}

// GetMemberService returns the member service
func (m *Module) GetMemberService() *bizservice.MemberService {
	return m.memberService
}

// GetBusinessHandler returns the business handler
func (m *Module) GetBusinessHandler() *bizhandler.BusinessHandler {
	return m.businessHandler
}

// GetMemberHandler returns the member handler
func (m *Module) GetMemberHandler() *bizhandler.MemberHandler {
	return m.memberHandler
}

func (m *Module) Init(ctx context.Context) error {
	log.Println("Business module initialized")
	return nil
}

func (m *Module) Close() {
	log.Println("Business module closed")
}