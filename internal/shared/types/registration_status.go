// internal/shared/types/registration_status.go

package types

// ============================================================
// REGISTRATION STATUS CONSTANTS
// ============================================================

// RegistrationStatusSlug constants (with hyphens) - Used for URLs and API routes
const (
	RegistrationStatusPendingSlug   = "pending"
	RegistrationStatusConfirmedSlug = "confirmed"
	RegistrationStatusCancelledSlug = "cancelled"
	RegistrationStatusRefundedSlug  = "refunded"
	RegistrationStatusExpiredSlug   = "expired"
	RegistrationStatusAttendedSlug  = "attended"
)

// RegistrationStatusName constants (with underscores) - Used for internal database lookups
const (
	RegistrationStatusPendingName   = "registration_status_pending"
	RegistrationStatusConfirmedName = "registration_status_confirmed"
	RegistrationStatusCancelledName = "registration_status_cancelled"
	RegistrationStatusRefundedName  = "registration_status_refunded"
	RegistrationStatusExpiredName   = "registration_status_expired"
	RegistrationStatusAttendedName  = "registration_status_attended"
)

// RegistrationStatusDisplayName constants - Used for UI display
const (
	RegistrationStatusPendingDisplayName   = "Pending"
	RegistrationStatusConfirmedDisplayName = "Confirmed"
	RegistrationStatusCancelledDisplayName = "Cancelled"
	RegistrationStatusRefundedDisplayName  = "Refunded"
	RegistrationStatusExpiredDisplayName   = "Expired"
	RegistrationStatusAttendedDisplayName  = "Attended"
)

// RegistrationStatusDescription constants
const (
	RegistrationStatusPendingDescription   = "Awaiting payment confirmation"
	RegistrationStatusConfirmedDescription = "Registration is active and confirmed"
	RegistrationStatusCancelledDescription = "Cancelled by user or organizer"
	RegistrationStatusRefundedDescription  = "Refunded after confirmation"
	RegistrationStatusExpiredDescription   = "Payment window elapsed without confirmation"
	RegistrationStatusAttendedDescription  = "Attendee attended the event"
)

// RegistrationStatusIcon constants
const (
	RegistrationStatusPendingIcon   = "clock"
	RegistrationStatusConfirmedIcon = "check-circle"
	RegistrationStatusCancelledIcon = "x-circle"
	RegistrationStatusRefundedIcon  = "rotate-ccw"
	RegistrationStatusExpiredIcon   = "alert-circle"
	RegistrationStatusAttendedIcon  = "user-check"
)

// RegistrationStatusColor constants
const (
	RegistrationStatusPendingColor   = "#F59E0B" // Amber
	RegistrationStatusConfirmedColor = "#10B981" // Green
	RegistrationStatusCancelledColor = "#EF4444" // Red
	RegistrationStatusRefundedColor  = "#8B5CF6" // Purple
	RegistrationStatusExpiredColor   = "#6B7280" // Gray
	RegistrationStatusAttendedColor  = "#3B82F6" // Blue
)

// RegistrationStatusSortOrder constants
const (
	RegistrationStatusPendingSortOrder   = 1
	RegistrationStatusConfirmedSortOrder = 2
	RegistrationStatusCancelledSortOrder = 3
	RegistrationStatusRefundedSortOrder  = 4
	RegistrationStatusExpiredSortOrder   = 5
	RegistrationStatusAttendedSortOrder  = 6
)

// ============================================================
// REGISTRATION STATUS DEFINITIONS
// ============================================================

// RegistrationStatus represents the lifecycle state of a registration.
//
// The canonical string value is the SLUG (e.g. "pending"), matching what
// appears in the registration_statuses.slug column and what the domain
// state machine uses.
type RegistrationStatus string

// RegistrationStatus constants — these use the SLUG as the canonical value
const (
	RegistrationStatusPending   RegistrationStatus = RegistrationStatusPendingSlug
	RegistrationStatusConfirmed RegistrationStatus = RegistrationStatusConfirmedSlug
	RegistrationStatusCancelled RegistrationStatus = RegistrationStatusCancelledSlug
	RegistrationStatusRefunded  RegistrationStatus = RegistrationStatusRefundedSlug
	RegistrationStatusExpired   RegistrationStatus = RegistrationStatusExpiredSlug
	RegistrationStatusAttended  RegistrationStatus = RegistrationStatusAttendedSlug
)

// AllRegistrationStatuses lists all valid statuses for validation
var AllRegistrationStatuses = []RegistrationStatus{
	RegistrationStatusPending,
	RegistrationStatusConfirmed,
	RegistrationStatusCancelled,
	RegistrationStatusRefunded,
	RegistrationStatusExpired,
	RegistrationStatusAttended,
}

// ============================================================
// BASIC METHODS
// ============================================================

// String returns the string representation (the slug)
func (s RegistrationStatus) String() string {
	return string(s)
}

// IsValid checks if the status is valid
func (s RegistrationStatus) IsValid() bool {
	for _, status := range AllRegistrationStatuses {
		if status == s {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS
// ============================================================

// GetSlug returns the slug (same as the canonical value)
func (s RegistrationStatus) GetSlug() string {
	return string(s)
}

// GetName returns the internal database name (with underscores prefix)
func (s RegistrationStatus) GetName() string {
	switch s {
	case RegistrationStatusPending:
		return RegistrationStatusPendingName
	case RegistrationStatusConfirmed:
		return RegistrationStatusConfirmedName
	case RegistrationStatusCancelled:
		return RegistrationStatusCancelledName
	case RegistrationStatusRefunded:
		return RegistrationStatusRefundedName
	case RegistrationStatusExpired:
		return RegistrationStatusExpiredName
	case RegistrationStatusAttended:
		return RegistrationStatusAttendedName
	default:
		return string(s)
	}
}

// GetDisplayName returns the user-facing display name - For UI
func (s RegistrationStatus) GetDisplayName() string {
	switch s {
	case RegistrationStatusPending:
		return RegistrationStatusPendingDisplayName
	case RegistrationStatusConfirmed:
		return RegistrationStatusConfirmedDisplayName
	case RegistrationStatusCancelled:
		return RegistrationStatusCancelledDisplayName
	case RegistrationStatusRefunded:
		return RegistrationStatusRefundedDisplayName
	case RegistrationStatusExpired:
		return RegistrationStatusExpiredDisplayName
	case RegistrationStatusAttended:
		return RegistrationStatusAttendedDisplayName
	default:
		return string(s)
	}
}

// GetDescription returns the description
func (s RegistrationStatus) GetDescription() string {
	switch s {
	case RegistrationStatusPending:
		return RegistrationStatusPendingDescription
	case RegistrationStatusConfirmed:
		return RegistrationStatusConfirmedDescription
	case RegistrationStatusCancelled:
		return RegistrationStatusCancelledDescription
	case RegistrationStatusRefunded:
		return RegistrationStatusRefundedDescription
	case RegistrationStatusExpired:
		return RegistrationStatusExpiredDescription
	case RegistrationStatusAttended:
		return RegistrationStatusAttendedDescription
	default:
		return ""
	}
}

// GetIcon returns the icon name for this status
func (s RegistrationStatus) GetIcon() string {
	switch s {
	case RegistrationStatusPending:
		return RegistrationStatusPendingIcon
	case RegistrationStatusConfirmed:
		return RegistrationStatusConfirmedIcon
	case RegistrationStatusCancelled:
		return RegistrationStatusCancelledIcon
	case RegistrationStatusRefunded:
		return RegistrationStatusRefundedIcon
	case RegistrationStatusExpired:
		return RegistrationStatusExpiredIcon
	case RegistrationStatusAttended:
		return RegistrationStatusAttendedIcon
	default:
		return "circle"
	}
}

// GetColor returns the color for this status
func (s RegistrationStatus) GetColor() string {
	switch s {
	case RegistrationStatusPending:
		return RegistrationStatusPendingColor
	case RegistrationStatusConfirmed:
		return RegistrationStatusConfirmedColor
	case RegistrationStatusCancelled:
		return RegistrationStatusCancelledColor
	case RegistrationStatusRefunded:
		return RegistrationStatusRefundedColor
	case RegistrationStatusExpired:
		return RegistrationStatusExpiredColor
	case RegistrationStatusAttended:
		return RegistrationStatusAttendedColor
	default:
		return "#6B7280"
	}
}

// GetSortOrder returns the sort order for this status
func (s RegistrationStatus) GetSortOrder() int {
	switch s {
	case RegistrationStatusPending:
		return RegistrationStatusPendingSortOrder
	case RegistrationStatusConfirmed:
		return RegistrationStatusConfirmedSortOrder
	case RegistrationStatusCancelled:
		return RegistrationStatusCancelledSortOrder
	case RegistrationStatusRefunded:
		return RegistrationStatusRefundedSortOrder
	case RegistrationStatusExpired:
		return RegistrationStatusExpiredSortOrder
	case RegistrationStatusAttended:
		return RegistrationStatusAttendedSortOrder
	default:
		return 999
	}
}

// ============================================================
// STATE MACHINE METHODS
// ============================================================

// IsFinal returns whether this status is terminal (no further transitions)
func (s RegistrationStatus) IsFinal() bool {
	switch s {
	case RegistrationStatusCancelled, RegistrationStatusRefunded, RegistrationStatusExpired:
		return true
	default:
		return false
	}
}

// IsActive returns whether a registration in this status counts toward capacity
func (s RegistrationStatus) IsActive() bool {
	return s == RegistrationStatusPending || s == RegistrationStatusConfirmed
}

// CanTransitionTo validates whether a status change is allowed.
func (s RegistrationStatus) CanTransitionTo(next RegistrationStatus) bool {
	switch s {
	case RegistrationStatusPending:
		return next == RegistrationStatusConfirmed ||
			next == RegistrationStatusExpired ||
			next == RegistrationStatusCancelled
	case RegistrationStatusConfirmed:
		return next == RegistrationStatusCancelled ||
			next == RegistrationStatusRefunded ||
			next == RegistrationStatusAttended
	case RegistrationStatusAttended:
		return next == RegistrationStatusRefunded
	case RegistrationStatusCancelled,
		RegistrationStatusRefunded,
		RegistrationStatusExpired:
		return false
	}
	return false
}

// ============================================================
// PARSE FUNCTIONS
// ============================================================

// ParseRegistrationStatus parses a string into a RegistrationStatus.
// Accepts the slug form ("pending").
func ParseRegistrationStatus(value string) (RegistrationStatus, bool) {
	s := RegistrationStatus(value)
	if s.IsValid() {
		return s, true
	}
	return "", false
}

// ParseRegistrationStatusWithDefault parses a string or returns a default
func ParseRegistrationStatusWithDefault(value string, defaultStatus RegistrationStatus) RegistrationStatus {
	if s, ok := ParseRegistrationStatus(value); ok {
		return s
	}
	return defaultStatus
}

// ParseRegistrationStatusBySlug is an alias for ParseRegistrationStatus
// kept for API symmetry with category parsing.
func ParseRegistrationStatusBySlug(slug string) (RegistrationStatus, bool) {
	return ParseRegistrationStatus(slug)
}

// ParseRegistrationStatusByName parses the internal name form
// (e.g. "registration_status_pending").
func ParseRegistrationStatusByName(name string) (RegistrationStatus, bool) {
	for _, s := range AllRegistrationStatuses {
		if s.GetName() == name {
			return s, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// AllRegistrationStatusNames returns all internal status names
// (with the "registration_status_" prefix).
func AllRegistrationStatusNames() []string {
	names := make([]string, 0, len(AllRegistrationStatuses))
	for _, s := range AllRegistrationStatuses {
		names = append(names, s.GetName())
	}
	return names
}

// AllRegistrationStatusSlugs returns all status slugs
func AllRegistrationStatusSlugs() []string {
	slugs := make([]string, 0, len(AllRegistrationStatuses))
	for _, s := range AllRegistrationStatuses {
		slugs = append(slugs, s.GetSlug())
	}
	return slugs
}

// AllRegistrationStatusDisplayNames returns all status display names
func AllRegistrationStatusDisplayNames() []string {
	names := make([]string, 0, len(AllRegistrationStatuses))
	for _, s := range AllRegistrationStatuses {
		names = append(names, s.GetDisplayName())
	}
	return names
}

// IsTerminalStatus checks if a status string represents a terminal state
func IsTerminalStatus(value string) bool {
	s, ok := ParseRegistrationStatus(value)
	if !ok {
		return false
	}
	return s.IsFinal()
}

// IsActiveStatus checks if a status string represents an active state
func IsActiveStatus(value string) bool {
	s, ok := ParseRegistrationStatus(value)
	if !ok {
		return false
	}
	return s.IsActive()
}