// internal/modules/registration/infrastructure/postgres/event_registration_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/registration/registrationdomain"
)

type EventRegistrationRepository struct {
	db *gorm.DB
}

func NewEventRegistrationRepository(db *gorm.DB) registrationdomain.EventRegistrationRepository {
	return &EventRegistrationRepository{db: db}
}

// ============================================================
// WRITE PATHS
// ============================================================

func (r *EventRegistrationRepository) Create(ctx context.Context, er *registrationdomain.EventRegistration) error {
	model := &EventRegistrationModel{
		RegistrationID: er.Registration.ID,
		EventID:        er.EventID,
		UserID:         nullableString(er.Registration.UserID),
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("create event registration: %w", err)
	}
	if err := r.replaceTickets(ctx, er.Registration.ID, er.Selections); err != nil {
		return err
	}
	return nil
}

func (r *EventRegistrationRepository) Update(ctx context.Context, er *registrationdomain.EventRegistration) error {
	res := r.db.WithContext(ctx).
		Model(&EventRegistrationModel{}).
		Where("registration_id = ?", er.Registration.ID).
		Updates(map[string]any{
			"event_id": er.EventID,
			"user_id":  nullableString(er.Registration.UserID),
		})
	if res.Error != nil {
		return fmt.Errorf("update event registration: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return registrationdomain.ErrRegistrationNotFound
	}
	if err := r.replaceTickets(ctx, er.Registration.ID, er.Selections); err != nil {
		return err
	}
	return nil
}

func (r *EventRegistrationRepository) FindByID(ctx context.Context, id string) (*registrationdomain.EventRegistration, error) {
	var model EventRegistrationModel
	err := r.db.WithContext(ctx).
		Preload("Registration").
		Preload("Registration.Status").
		Where("registration_id = ?", id).
		First(&model).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, registrationdomain.ErrRegistrationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find event registration: %w", err)
	}

	tickets, err := r.loadTickets(ctx, id)
	if err != nil {
		return nil, err
	}
	return toEventRegistrationDomain(&model, tickets)
}

func (r *EventRegistrationRepository) ListByEvent(ctx context.Context, eventID string, f registrationdomain.ListFilter) ([]*registrationdomain.EventRegistration, int, error) {
	query := r.db.WithContext(ctx).
		Model(&EventRegistrationModel{}).
		Where("event_id = ?", eventID)

	if len(f.Statuses) > 0 {
		slugs := statusesToSlugs(f.Statuses)
		query = query.
			Joins("JOIN registrations r ON r.id = event_registrations.registration_id").
			Joins("JOIN registration_statuses rs ON rs.id = r.status_id").
			Where("rs.slug IN ?", slugs)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count event registrations: %w", err)
	}

	var models []EventRegistrationModel
	err := query.
		Preload("Registration").
		Preload("Registration.Status").
		Order("event_registrations.created_at DESC").
		Limit(f.PageSize).
		Offset((f.Page - 1) * f.PageSize).
		Find(&models).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list event registrations: %w", err)
	}

	out := make([]*registrationdomain.EventRegistration, 0, len(models))
	for i := range models {
		tickets, err := r.loadTickets(ctx, models[i].RegistrationID)
		if err != nil {
			return nil, 0, err
		}
		er, err := toEventRegistrationDomain(&models[i], tickets)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, er)
	}
	return out, int(total), nil
}

func (r *EventRegistrationRepository) ListByUser(ctx context.Context, userID string, f registrationdomain.ListFilter) ([]*registrationdomain.EventRegistration, int, error) {
	query := r.db.WithContext(ctx).
		Model(&EventRegistrationModel{}).
		Joins("JOIN registrations r ON r.id = event_registrations.registration_id").
		Where("r.user_id = ?", userID)

	if len(f.Statuses) > 0 {
		slugs := statusesToSlugs(f.Statuses)
		query = query.
			Joins("JOIN registration_statuses rs ON rs.id = r.status_id").
			Where("rs.slug IN ?", slugs)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count user registrations: %w", err)
	}

	var models []EventRegistrationModel
	err := query.
		Preload("Registration").
		Preload("Registration.Status").
		Order("event_registrations.created_at DESC").
		Limit(f.PageSize).
		Offset((f.Page - 1) * f.PageSize).
		Find(&models).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list user registrations: %w", err)
	}

	out := make([]*registrationdomain.EventRegistration, 0, len(models))
	for i := range models {
		tickets, err := r.loadTickets(ctx, models[i].RegistrationID)
		if err != nil {
			return nil, 0, err
		}
		er, err := toEventRegistrationDomain(&models[i], tickets)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, er)
	}
	return out, int(total), nil
}

// ============================================================
// HELPERS
// ============================================================

func (r *EventRegistrationRepository) replaceTickets(ctx context.Context, registrationID string, selections []registrationdomain.TicketSelection) error {
	if err := r.db.WithContext(ctx).
		Where("registration_id = ?", registrationID).
		Delete(&EventRegistrationTicketModel{}).Error; err != nil {
		return fmt.Errorf("delete tickets: %w", err)
	}

	models := make([]EventRegistrationTicketModel, 0, len(selections))
	for _, s := range selections {
		models = append(models, EventRegistrationTicketModel{
			RegistrationID: registrationID,
			TicketTypeID:   s.TicketTypeID,
			Quantity:       s.Quantity,
			UnitPrice:      s.UnitPrice,
			Discount:       s.Discount,
		})
	}
	if len(models) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&models).Error; err != nil {
		return fmt.Errorf("insert tickets: %w", err)
	}
	return nil
}

func (r *EventRegistrationRepository) loadTickets(ctx context.Context, registrationID string) ([]registrationdomain.TicketSelection, error) {
	var models []EventRegistrationTicketModel
	err := r.db.WithContext(ctx).
		Where("registration_id = ?", registrationID).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("load tickets: %w", err)
	}

	out := make([]registrationdomain.TicketSelection, 0, len(models))
	for _, m := range models {
		out = append(out, registrationdomain.TicketSelection{
			TicketTypeID: m.TicketTypeID,
			Quantity:     m.Quantity,
			UnitPrice:    m.UnitPrice,
			Discount:     m.Discount,
		})
	}
	return out, nil
}

func statusesToSlugs(statuses []registrationdomain.Status) []string {
	slugs := make([]string, 0, len(statuses))
	for _, s := range statuses {
		slugs = append(slugs, s.GetSlug())
	}
	return slugs
}

func (r *EventRegistrationRepository) Deactivate(
	ctx context.Context,
	registrationID string,
) error {
	return r.db.WithContext(ctx).
		Model(&EventRegistrationModel{}).
		Where("registration_id = ?", registrationID).
		Update("is_active", false).Error
}

// ============================================================
// CROSS-EVENT LIST (organizer view)
// ============================================================
//
// Implementation note: we use raw SQL here because GORM's
// Table().Select(raw).Scan() chain silently dropped aliased
// columns and returned zero rows on this codebase. Raw SQL
// guarantees the exact query we write is what Postgres runs.
// ============================================================

type crossEventRow struct {
	ID                 string    `gorm:"column:id"`
	RegistrationNumber string    `gorm:"column:registration_number"`
	StatusSlug         string    `gorm:"column:status_slug"`
	UserID             string    `gorm:"column:user_id"`
	GuestName          string    `gorm:"column:guest_name"`
	GuestEmail         string    `gorm:"column:guest_email"`
	GuestPhone         string    `gorm:"column:guest_phone"`
	UserName           string    `gorm:"column:user_name"`
	UserEmail          string    `gorm:"column:user_email"`
	EventID            string    `gorm:"column:event_id"`
	EventName          string    `gorm:"column:event_name"`
	EventStartDate     time.Time `gorm:"column:event_start_date"`
	EventImageURL      string    `gorm:"column:event_image_url"`
	IsVirtual          bool      `gorm:"column:is_virtual"`
	IsHybrid           bool      `gorm:"column:is_hybrid"`
	VenueName          string    `gorm:"column:venue_name"`
	VenueAddress       string    `gorm:"column:venue_address"`
	VenueCity          string    `gorm:"column:venue_city"`
	VenueCountry       string    `gorm:"column:venue_country"`
	InPersonLocation   string    `gorm:"column:in_person_location"`
	TicketName         string    `gorm:"column:ticket_name"`
	CreatedAt          time.Time `gorm:"column:created_at"`
}

func (r *EventRegistrationRepository) ListAll(
	ctx context.Context,
	f registrationdomain.ListAllFilter,
) ([]*registrationdomain.CrossEventRegistrationRow, int, error) {

	if len(f.AccountIDs) == 0 {
		return []*registrationdomain.CrossEventRegistrationRow{}, 0, nil
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// ------------------------------------------------------------
	// WHERE fragments + args
	// ------------------------------------------------------------
	where := []string{
		"e.team_id IN (SELECT t.id FROM teams t WHERE t.account_id IN ? AND t.deleted_at IS NULL)",
	}
	args := []any{f.AccountIDs}

	if f.EventID != "" {
		where = append(where, "er.event_id = ?")
		args = append(args, f.EventID)
	}

	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		where = append(where,
			"(LOWER(COALESCE(r.guest_name, '')) LIKE ? "+
				"OR LOWER(COALESCE(r.guest_email, '')) LIKE ? "+
				"OR LOWER(COALESCE(u.name, '')) LIKE ? "+
				"OR LOWER(COALESCE(u.email, '')) LIKE ?)")
		args = append(args, like, like, like, like)
	}

	if len(f.Statuses) > 0 {
		slugs := statusesToSlugs(f.Statuses)
		where = append(where, "rs.slug IN ?")
		args = append(args, slugs)
	}

	whereSQL := strings.Join(where, " AND ")

	// ------------------------------------------------------------
	// COUNT
	// ------------------------------------------------------------
	countSQL := `
		SELECT COUNT(*)
		FROM event_registrations AS er
		JOIN registrations r        ON r.id = er.registration_id AND r.deleted_at IS NULL
		JOIN events e               ON e.id = er.event_id        AND e.deleted_at IS NULL
		JOIN registration_statuses rs ON rs.id = r.status_id
		LEFT JOIN users u           ON u.id = r.user_id          AND u.deleted_at IS NULL
		WHERE ` + whereSQL

	var total int64
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count all registrations: %w", err)
	}

	// ------------------------------------------------------------
	// ORDER
	// ------------------------------------------------------------
	orderClause := buildListAllOrderClause(f.SortBy, f.SortOrder)

	// ------------------------------------------------------------
	// DATA
	// ------------------------------------------------------------
	dataSQL := `
		SELECT
			r.id,
			r.registration_number,
			rs.slug AS status_slug,
			COALESCE(r.user_id::text, '') AS user_id,
			COALESCE(r.guest_name, '')    AS guest_name,
			COALESCE(r.guest_email, '')   AS guest_email,
			COALESCE(r.guest_phone, '')   AS guest_phone,
			COALESCE(u.name, '')          AS user_name,
			COALESCE(u.email, '')         AS user_email,
			er.event_id,
			e.display_name                AS event_name,
			e.start_date                  AS event_start_date,
			COALESCE(e.image_url, '') AS event_image_url,
			COALESCE(e.is_virtual, false) AS is_virtual,
			COALESCE(e.is_hybrid, false)  AS is_hybrid,
			COALESCE(e.venue_name, '')          AS venue_name,
			COALESCE(e.venue_address, '')       AS venue_address,
			COALESCE(e.venue_city, '')          AS venue_city,
			COALESCE(e.venue_country, '')       AS venue_country,
			COALESCE(e.in_person_location, '')  AS in_person_location,
			COALESCE(tt.display_name, '') AS ticket_name,
			r.created_at
		FROM event_registrations AS er
		JOIN registrations r        ON r.id = er.registration_id AND r.deleted_at IS NULL
		JOIN events e               ON e.id = er.event_id        AND e.deleted_at IS NULL
		JOIN registration_statuses rs ON rs.id = r.status_id
		LEFT JOIN users u           ON u.id = r.user_id          AND u.deleted_at IS NULL
		LEFT JOIN LATERAL (
			SELECT ert.ticket_type_id
			FROM event_registration_tickets ert
			WHERE ert.registration_id = r.id
			ORDER BY ert.ticket_type_id
			LIMIT 1
		) AS ert ON true
		LEFT JOIN ticket_types tt   ON tt.id = ert.ticket_type_id AND tt.deleted_at IS NULL
		WHERE ` + whereSQL + `
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`
	dataArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)

	var rows []crossEventRow
	if err := r.db.WithContext(ctx).Raw(dataSQL, dataArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("list all registrations: %w", err)
	}

	out := make([]*registrationdomain.CrossEventRegistrationRow, 0, len(rows))
	for _, x := range rows {
		isGuest := x.UserID == ""
		attendeeName := x.GuestName
		email := x.GuestEmail
		phone := x.GuestPhone
		if !isGuest {
			if x.UserName != "" {
				attendeeName = x.UserName
			}
			if x.UserEmail != "" {
				email = x.UserEmail
			}
		}

		out = append(out, &registrationdomain.CrossEventRegistrationRow{
			ID:                 x.ID,
			RegistrationNumber: x.RegistrationNumber,
			Status:             registrationdomain.Status(x.StatusSlug),
			UserID:             x.UserID,
			AttendeeName:       attendeeName,
			Email:              email,
			Phone:              phone,
			IsGuest:            isGuest,
			EventID:            x.EventID,
			EventName:          x.EventName,
			EventStartDate:     x.EventStartDate,
			EventImageURL:      x.EventImageURL, 
			IsVirtual:          x.IsVirtual,
			IsHybrid:           x.IsHybrid,
			VenueName:          x.VenueName,
			VenueAddress:       x.VenueAddress,
			VenueCity:          x.VenueCity,
			VenueCountry:       x.VenueCountry,
			InPersonLocation:   x.InPersonLocation,
			TicketName:         x.TicketName,
			CreatedAt:          x.CreatedAt,
		})
	}

	return out, int(total), nil
}

func buildListAllOrderClause(sortBy, sortOrder string) string {
	dir := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		dir = "ASC"
	}
	switch sortBy {
	case "attendee_name":
		return "COALESCE(NULLIF(u.name, ''), r.guest_name, '') " + dir
	case "event_name":
		return "e.display_name " + dir
	case "status":
		return "rs.slug " + dir
	case "created_at", "":
		return "r.created_at " + dir
	default:
		return "r.created_at DESC"
	}
}

// ============================================================
// CROSS-EVENT LIST (attendee view)
// ============================================================

func (r *EventRegistrationRepository) ListMine(
	ctx context.Context,
	f registrationdomain.ListMineFilter,
) ([]*registrationdomain.CrossEventRegistrationRow, int, error) {

	if f.UserID == "" {
		return []*registrationdomain.CrossEventRegistrationRow{}, 0, nil
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	where := []string{"r.user_id = ?"}
	args := []any{f.UserID}

	if f.EventID != "" {
		where = append(where, "er.event_id = ?")
		args = append(args, f.EventID)
	}

	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		where = append(where,
			"(LOWER(COALESCE(e.display_name, '')) LIKE ? "+
				"OR LOWER(COALESCE(e.name, '')) LIKE ?)")
		args = append(args, like, like)
	}

	if len(f.Statuses) > 0 {
		slugs := statusesToSlugs(f.Statuses)
		where = append(where, "rs.slug IN ?")
		args = append(args, slugs)
	}

	whereSQL := strings.Join(where, " AND ")

	// ------------------------------------------------------------
	// COUNT
	// ------------------------------------------------------------
	countSQL := `
		SELECT COUNT(*)
		FROM event_registrations AS er
		JOIN registrations r        ON r.id = er.registration_id AND r.deleted_at IS NULL
		JOIN events e               ON e.id = er.event_id        AND e.deleted_at IS NULL
		JOIN registration_statuses rs ON rs.id = r.status_id
		WHERE ` + whereSQL

	var total int64
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count my registrations: %w", err)
	}

	orderClause := buildListMineOrderClause(f.SortBy, f.SortOrder)

	// ------------------------------------------------------------
	// DATA
	// ------------------------------------------------------------
	dataSQL := `
		SELECT
			r.id,
			r.registration_number,
			rs.slug AS status_slug,
			COALESCE(r.user_id::text, '') AS user_id,
			COALESCE(r.guest_name, '')    AS guest_name,
			COALESCE(r.guest_email, '')   AS guest_email,
			COALESCE(r.guest_phone, '')   AS guest_phone,
			COALESCE(u.name, '')          AS user_name,
			COALESCE(u.email, '')         AS user_email,
			er.event_id,
			e.display_name                AS event_name,
			e.start_date                  AS event_start_date,
			COALESCE(e.image_url, '') AS event_image_url,
			COALESCE(e.is_virtual, false) AS is_virtual,
			COALESCE(e.is_hybrid, false)  AS is_hybrid,
			COALESCE(e.venue_name, '')          AS venue_name,
			COALESCE(e.venue_address, '')       AS venue_address,
			COALESCE(e.venue_city, '')          AS venue_city,
			COALESCE(e.venue_country, '')       AS venue_country,
			COALESCE(e.in_person_location, '')  AS in_person_location,
			COALESCE(tt.display_name, '') AS ticket_name,
			r.created_at
		FROM event_registrations AS er
		JOIN registrations r        ON r.id = er.registration_id AND r.deleted_at IS NULL
		JOIN events e               ON e.id = er.event_id        AND e.deleted_at IS NULL
		JOIN registration_statuses rs ON rs.id = r.status_id
		LEFT JOIN users u           ON u.id = r.user_id          AND u.deleted_at IS NULL
		LEFT JOIN LATERAL (
			SELECT ert.ticket_type_id
			FROM event_registration_tickets ert
			WHERE ert.registration_id = r.id
			ORDER BY ert.ticket_type_id
			LIMIT 1
		) AS ert ON true
		LEFT JOIN ticket_types tt   ON tt.id = ert.ticket_type_id AND tt.deleted_at IS NULL
		WHERE ` + whereSQL + `
		ORDER BY ` + orderClause + `
		LIMIT ? OFFSET ?
	`
	dataArgs := append(append([]any{}, args...), pageSize, (page-1)*pageSize)

	var rows []crossEventRow
	if err := r.db.WithContext(ctx).Raw(dataSQL, dataArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("list my registrations: %w", err)
	}

	out := make([]*registrationdomain.CrossEventRegistrationRow, 0, len(rows))
	for _, x := range rows {
		isGuest := x.UserID == ""
		attendeeName := x.GuestName
		email := x.GuestEmail
		phone := x.GuestPhone
		if !isGuest {
			if x.UserName != "" {
				attendeeName = x.UserName
			}
			if x.UserEmail != "" {
				email = x.UserEmail
			}
			phone = ""
		}

		out = append(out, &registrationdomain.CrossEventRegistrationRow{
			ID:                 x.ID,
			RegistrationNumber: x.RegistrationNumber,
			Status:             registrationdomain.Status(x.StatusSlug),
			UserID:             x.UserID,
			AttendeeName:       attendeeName,
			Email:              email,
			Phone:              phone,
			IsGuest:            isGuest,
			EventID:            x.EventID,
			EventName:          x.EventName,
			EventStartDate:     x.EventStartDate,
			EventImageURL:      x.EventImageURL, 
			IsVirtual:          x.IsVirtual,
			IsHybrid:           x.IsHybrid,
			VenueName:          x.VenueName,
			VenueAddress:       x.VenueAddress,
			VenueCity:          x.VenueCity,
			VenueCountry:       x.VenueCountry,
			InPersonLocation:   x.InPersonLocation,
			TicketName:         x.TicketName,
			CreatedAt:          x.CreatedAt,
		})
	}

	return out, int(total), nil
}


func buildListMineOrderClause(sortBy, sortOrder string) string {
	dir := "DESC"
	if strings.EqualFold(sortOrder, "asc") {
		dir = "ASC"
	}
	switch sortBy {
	case "event_name":
		return "e.display_name " + dir
	case "status":
		return "rs.slug " + dir
	case "created_at", "":
		return "r.created_at " + dir
	default:
		return "r.created_at DESC"
	}
}
