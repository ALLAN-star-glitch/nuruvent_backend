package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	attendance "github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/attendance/attendancedomain"
)

type AttendeeRollupStatusRepository struct {
	db *gorm.DB
}

func NewAttendeeRollupStatusRepository(db *gorm.DB) *AttendeeRollupStatusRepository {
	return &AttendeeRollupStatusRepository{db: db}
}

func (r *AttendeeRollupStatusRepository) Upsert(
	ctx context.Context,
	s *attendance.AttendeeRollupStatus,
) error {
	model := toRollupStatusModel(s)
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "attendee_id"},
				{Name: "external_type"},
				{Name: "external_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"derived_status",
				"sessions_total",
				"sessions_attended",
				"sessions_confirmed",
				"total_duration_seconds",
				"last_derived_at",
			}),
		}).
		Create(model).Error
	if err != nil {
		return translateError(err, "upsert rollup status", attendance.ErrStatusNotFound)
	}
	return nil
}

func (r *AttendeeRollupStatusRepository) FindByAttendeeExternal(
	ctx context.Context,
	attendeeID string,
	ref attendance.ExternalRef,
) (*attendance.AttendeeRollupStatus, error) {
	var m AttendeeRollupStatusModel
	err := r.db.WithContext(ctx).
		Where("attendee_id = ? AND external_type = ? AND external_id = ?",
			attendeeID, ref.Type, ref.ID).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrStatusNotFound
	}
	if err != nil {
		return nil, translateError(err, "find rollup status", attendance.ErrStatusNotFound)
	}
	return toRollupStatusDomain(&m), nil
}

func (r *AttendeeRollupStatusRepository) ListByExternal(
	ctx context.Context,
	ref attendance.ExternalRef,
) ([]*attendance.AttendeeRollupStatus, error) {
	var models []AttendeeRollupStatusModel
	err := r.db.WithContext(ctx).
		Where("external_type = ? AND external_id = ?", ref.Type, ref.ID).
		Find(&models).Error
	if err != nil {
		return nil, translateError(err, "list rollup statuses", attendance.ErrStatusNotFound)
	}
	out := make([]*attendance.AttendeeRollupStatus, 0, len(models))
	for i := range models {
		out = append(out, toRollupStatusDomain(&models[i]))
	}
	return out, nil
}

// ============================================================
// EVENT ATTENDEE LIST + DETAIL
// ============================================================

const (
	defaultAttendeePageSize = 20
	maxAttendeePageSize     = 100
)

// ListEventAttendees returns a paginated attendee list for an event.
// It reads the pre-computed attendee_rollup_statuses table and joins
// attendees for display fields. No aggregation happens here — the
// rollup is maintained by RecomputeRollup.
func (r *AttendeeRollupStatusRepository) ListEventAttendees(
	ctx context.Context,
	q attendance.ListEventAttendeesQuery,
) (*attendance.ListEventAttendeesResult, error) {

	// --- Normalize paging ---
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = defaultAttendeePageSize
	}
	if pageSize > maxAttendeePageSize {
		pageSize = maxAttendeePageSize
	}
	offset := (page - 1) * pageSize

	// --- Base query ---
	base := r.db.WithContext(ctx).
		Table("attendee_rollup_statuses AS ars").
		Joins("JOIN attendees a ON a.id = ars.attendee_id AND a.deleted_at IS NULL").
		Where("ars.external_type = ? AND ars.external_id = ?",
			"event", q.EventID)

	// --- Search filter ---
	if s := strings.TrimSpace(q.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		base = base.Where(
			"(LOWER(a.display_name) LIKE ? OR LOWER(a.email) LIKE ?)",
			like, like,
		)
	}

	// --- Status filter ---
	if len(q.Statuses) > 0 {
		base = base.Where("ars.derived_status IN ?", q.Statuses)
	}

	// --- Count (before ordering/limit) ---
	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, translateError(err, "count event attendees", attendance.ErrStatusNotFound)
	}

	// --- Order ---
	orderClause := buildAttendeeOrderClause(q.SortBy, q.SortOrder)

	// --- Rows ---
	type row struct {
		AttendeeID           string    `gorm:"column:attendee_id"`
		DisplayName          string    `gorm:"column:display_name"`
		Email                string    `gorm:"column:email"`
		DerivedStatus        string    `gorm:"column:derived_status"`
		SessionsTotal        int       `gorm:"column:sessions_total"`
		SessionsAttended     int       `gorm:"column:sessions_attended"`
		SessionsConfirmed    int       `gorm:"column:sessions_confirmed"`
		TotalDurationSeconds int       `gorm:"column:total_duration_seconds"`
		LastDerivedAt        time.Time `gorm:"column:last_derived_at"`
		RegisteredAt         time.Time `gorm:"column:registered_at"`
	}

	var rows []row
	err := base.
		Select(
			"ars.attendee_id, " +
				"a.display_name, " +
				"a.email, " +
				"ars.derived_status, " +
				"ars.sessions_total, " +
				"ars.sessions_attended, " +
				"ars.sessions_confirmed, " +
				"ars.total_duration_seconds, " +
				"ars.last_derived_at, " +
				"a.created_at AS registered_at",
		).
		Order(orderClause).
		Limit(pageSize).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, translateError(err, "list event attendees", attendance.ErrStatusNotFound)
	}

	out := make([]*attendance.EventAttendeeRow, 0, len(rows))
	for _, x := range rows {
		out = append(out, &attendance.EventAttendeeRow{
			AttendeeID:           x.AttendeeID,
			DisplayName:          x.DisplayName,
			Email:                x.Email,
			DerivedStatus:        x.DerivedStatus,
			SessionsTotal:        x.SessionsTotal,
			SessionsAttended:     x.SessionsAttended,
			SessionsConfirmed:    x.SessionsConfirmed,
			TotalDurationSeconds: x.TotalDurationSeconds,
			LastDerivedAt:        x.LastDerivedAt,
			RegisteredAt:         x.RegisteredAt,
		})
	}

	return &attendance.ListEventAttendeesResult{
		Attendees: out,
		Total:     int(total),
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// FindEventAttendee returns a single attendee's rollup for an event,
// joined with the attendee record for display fields.
func (r *AttendeeRollupStatusRepository) FindEventAttendee(
	ctx context.Context,
	eventID, attendeeID string,
) (*attendance.EventAttendeeRow, error) {

	type row struct {
		AttendeeID           string    `gorm:"column:attendee_id"`
		DisplayName          string    `gorm:"column:display_name"`
		Email                string    `gorm:"column:email"`
		DerivedStatus        string    `gorm:"column:derived_status"`
		SessionsTotal        int       `gorm:"column:sessions_total"`
		SessionsAttended     int       `gorm:"column:sessions_attended"`
		SessionsConfirmed    int       `gorm:"column:sessions_confirmed"`
		TotalDurationSeconds int       `gorm:"column:total_duration_seconds"`
		LastDerivedAt        time.Time `gorm:"column:last_derived_at"`
		RegisteredAt         time.Time `gorm:"column:registered_at"`
	}

	var x row
	err := r.db.WithContext(ctx).
		Table("attendee_rollup_statuses AS ars").
		Joins("JOIN attendees a ON a.id = ars.attendee_id AND a.deleted_at IS NULL").
		Where("ars.external_type = ? AND ars.external_id = ? AND ars.attendee_id = ?",
			"event", eventID, attendeeID).
		Select(
			"ars.attendee_id, " +
				"a.display_name, " +
				"a.email, " +
				"ars.derived_status, " +
				"ars.sessions_total, " +
				"ars.sessions_attended, " +
				"ars.sessions_confirmed, " +
				"ars.total_duration_seconds, " +
				"ars.last_derived_at, " +
				"a.created_at AS registered_at",
		).
		Take(&x).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, attendance.ErrAttendeeNotFound
	}
	if err != nil {
		return nil, translateError(err, "find event attendee", attendance.ErrAttendeeNotFound)
	}

	return &attendance.EventAttendeeRow{
		AttendeeID:           x.AttendeeID,
		DisplayName:          x.DisplayName,
		Email:                x.Email,
		DerivedStatus:        x.DerivedStatus,
		SessionsTotal:        x.SessionsTotal,
		SessionsAttended:     x.SessionsAttended,
		SessionsConfirmed:    x.SessionsConfirmed,
		TotalDurationSeconds: x.TotalDurationSeconds,
		LastDerivedAt:        x.LastDerivedAt,
		RegisteredAt:         x.RegisteredAt,
	}, nil
}


// ListAttendees returns a paginated list of attendees across every
// event owned by any of the caller's accounts.
//
// Ownership chain: events.team_id → teams.account_id → account in
// AccountIDs.
func (r *AttendeeRollupStatusRepository) ListAttendees(
	ctx context.Context,
	q attendance.ListAttendeesQuery,
) (*attendance.ListAttendeesResult, error) {

	if len(q.AccountIDs) == 0 {
		// No accounts → no rows. Refuse to fall through to a full
		// table scan.
		return &attendance.ListAttendeesResult{
			Attendees: []*attendance.CrossEventAttendeeRow{},
			Total:     0,
			Page:      maxInt(q.Page, 1),
			PageSize:  defaultAttendeePageSize,
		}, nil
	}

	// --- Normalize paging ---
	page := q.Page
	if page < 1 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = defaultAttendeePageSize
	}
	if pageSize > maxAttendeePageSize {
		pageSize = maxAttendeePageSize
	}
	offset := (page - 1) * pageSize

	// --- Base query ---
	base := r.db.WithContext(ctx).
		Table("attendee_rollup_statuses AS ars").
		Joins("JOIN attendees a ON a.id = ars.attendee_id AND a.deleted_at IS NULL").
		Joins("JOIN events e ON e.id = ars.external_id AND e.deleted_at IS NULL").
		Where("ars.external_type = ?", "event").
		Where(
			"e.team_id IN (SELECT t.id FROM teams t WHERE t.account_id IN ? AND t.deleted_at IS NULL)",
			q.AccountIDs,
		)

	// --- Event filter ---
	if q.EventID != "" {
		base = base.Where("ars.external_id = ?", q.EventID)
	}

	// --- Search ---
	if s := strings.TrimSpace(q.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		base = base.Where(
			"(LOWER(a.display_name) LIKE ? OR LOWER(a.email) LIKE ?)",
			like, like,
		)
	}

	// --- Status filter ---
	if len(q.Statuses) > 0 {
		base = base.Where("ars.derived_status IN ?", q.Statuses)
	}

	// --- Count ---
	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, translateError(err, "count attendees", attendance.ErrStatusNotFound)
	}

	// --- Order ---
	orderClause := buildCrossEventAttendeeOrderClause(q.SortBy, q.SortOrder)

	// --- Rows ---
	type row struct {
		AttendeeID           string    `gorm:"column:attendee_id"`
		DisplayName          string    `gorm:"column:display_name"`
		Email                string    `gorm:"column:email"`
		EventID              string    `gorm:"column:event_id"`
		EventName            string    `gorm:"column:event_name"`
		EventSlug            string    `gorm:"column:event_slug"`
		EventStartDate       time.Time `gorm:"column:event_start_date"`
		DerivedStatus        string    `gorm:"column:derived_status"`
		SessionsTotal        int       `gorm:"column:sessions_total"`
		SessionsAttended     int       `gorm:"column:sessions_attended"`
		SessionsConfirmed    int       `gorm:"column:sessions_confirmed"`
		TotalDurationSeconds int64     `gorm:"column:total_duration_seconds"`
		LastDerivedAt        time.Time `gorm:"column:last_derived_at"`
		RegisteredAt         time.Time `gorm:"column:registered_at"`
	}

	var rows []row
	err := base.
		Select(
			"ars.attendee_id, " +
				"a.display_name, " +
				"a.email, " +
				"ars.external_id AS event_id, " +
				"e.display_name AS event_name, " +
				"e.slug AS event_slug, " +
				"e.start_date AS event_start_date, " +
				"ars.derived_status, " +
				"ars.sessions_total, " +
				"ars.sessions_attended, " +
				"ars.sessions_confirmed, " +
				"ars.total_duration_seconds, " +
				"ars.last_derived_at, " +
				"a.created_at AS registered_at",
		).
		Order(orderClause).
		Limit(pageSize).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, translateError(err, "list attendees", attendance.ErrStatusNotFound)
	}

	out := make([]*attendance.CrossEventAttendeeRow, 0, len(rows))
	for _, x := range rows {
		out = append(out, &attendance.CrossEventAttendeeRow{
			AttendeeID:           x.AttendeeID,
			DisplayName:          x.DisplayName,
			Email:                x.Email,
			EventID:              x.EventID,
			EventName:            x.EventName,
			EventSlug:            x.EventSlug,
			EventStartDate:       x.EventStartDate,
			DerivedStatus:        x.DerivedStatus,
			SessionsTotal:        x.SessionsTotal,
			SessionsAttended:     x.SessionsAttended,
			SessionsConfirmed:    x.SessionsConfirmed,
			TotalDurationSeconds: x.TotalDurationSeconds,
			LastDerivedAt:        x.LastDerivedAt,
			RegisteredAt:         x.RegisteredAt,
		})
	}

	return &attendance.ListAttendeesResult{
		Attendees: out,
		Total:     int(total),
		Page:      page,
		PageSize:  pageSize,
	}, nil
}

// buildCrossEventAttendeeOrderClause maps the API sort field to a
// safe SQL clause.
func buildCrossEventAttendeeOrderClause(sortBy, sortOrder string) string {
	dir := "ASC"
	if strings.EqualFold(sortOrder, "desc") {
		dir = "DESC"
	}

	switch sortBy {
	case "name":
		return "a.display_name " + dir
	case "event":
		return "e.display_name " + dir + ", a.display_name ASC"
	case "registered_at":
		return "a.created_at " + dir
	case "status":
		return "ars.derived_status " + dir
	case "duration":
		return "ars.total_duration_seconds " + dir
	default:
		return "a.created_at DESC"
	}
}

// maxInt is a small helper for the empty-account early return.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// buildAttendeeOrderClause maps the API sort field to a safe SQL
// clause. Never interpolate user input directly.
func buildAttendeeOrderClause(sortBy, sortOrder string) string {
	dir := "ASC"
	if strings.EqualFold(sortOrder, "desc") {
		dir = "DESC"
	}

	switch sortBy {
	case "name":
		return "a.display_name " + dir
	case "registered_at":
		return "a.created_at " + dir
	case "status":
		return "ars.derived_status " + dir
	case "duration":
		return "ars.total_duration_seconds " + dir
	default:
		return "a.display_name ASC"
	}
}

// compile-time assertion
var _ attendance.AttendeeRollupStatusRepository = (*AttendeeRollupStatusRepository)(nil)