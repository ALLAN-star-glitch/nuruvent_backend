// internal/shared/types/waitlist_status.go

package types

// ============================================================
// WAITLIST STATUS CONSTANTS
// ============================================================

const (
	WaitlistStatusWaitingSlug   = "waiting"
	WaitlistStatusOfferedSlug   = "offered"
	WaitlistStatusConvertedSlug = "converted"
	WaitlistStatusExpiredSlug   = "expired"
	WaitlistStatusCancelledSlug = "cancelled"
)

const (
	WaitlistStatusWaitingName   = "waitlist_status_waiting"
	WaitlistStatusOfferedName   = "waitlist_status_offered"
	WaitlistStatusConvertedName = "waitlist_status_converted"
	WaitlistStatusExpiredName   = "waitlist_status_expired"
	WaitlistStatusCancelledName = "waitlist_status_cancelled"
)

const (
	WaitlistStatusWaitingDisplayName   = "Waiting"
	WaitlistStatusOfferedDisplayName   = "Offered"
	WaitlistStatusConvertedDisplayName = "Converted"
	WaitlistStatusExpiredDisplayName   = "Expired"
	WaitlistStatusCancelledDisplayName = "Cancelled"
)

const (
	WaitlistStatusWaitingDescription   = "Waiting for a spot to open"
	WaitlistStatusOfferedDescription   = "A spot was offered, awaiting acceptance"
	WaitlistStatusConvertedDescription = "Converted into a registration"
	WaitlistStatusExpiredDescription   = "Offer expired before acceptance"
	WaitlistStatusCancelledDescription = "Cancelled by user or organizer"
)

const (
	WaitlistStatusWaitingIcon   = "clock"
	WaitlistStatusOfferedIcon   = "bell"
	WaitlistStatusConvertedIcon = "check-circle"
	WaitlistStatusExpiredIcon   = "x-circle"
	WaitlistStatusCancelledIcon = "x-circle"
)

const (
	WaitlistStatusWaitingColor   = "#F59E0B" // Amber
	WaitlistStatusOfferedColor   = "#3B82F6" // Blue
	WaitlistStatusConvertedColor = "#10B981" // Green
	WaitlistStatusExpiredColor   = "#6B7280" // Gray
	WaitlistStatusCancelledColor = "#EF4444" // Red
)

const (
	WaitlistStatusWaitingSortOrder   = 1
	WaitlistStatusOfferedSortOrder   = 2
	WaitlistStatusConvertedSortOrder = 3
	WaitlistStatusExpiredSortOrder   = 4
	WaitlistStatusCancelledSortOrder = 5
)

// ============================================================
// DEFINITIONS
// ============================================================

type WaitlistStatus string

const (
	WaitlistStatusWaiting   WaitlistStatus = WaitlistStatusWaitingSlug
	WaitlistStatusOffered   WaitlistStatus = WaitlistStatusOfferedSlug
	WaitlistStatusConverted WaitlistStatus = WaitlistStatusConvertedSlug
	WaitlistStatusExpired   WaitlistStatus = WaitlistStatusExpiredSlug
	WaitlistStatusCancelled WaitlistStatus = WaitlistStatusCancelledSlug
)

var AllWaitlistStatuses = []WaitlistStatus{
	WaitlistStatusWaiting,
	WaitlistStatusOffered,
	WaitlistStatusConverted,
	WaitlistStatusExpired,
	WaitlistStatusCancelled,
}

// ============================================================
// METHODS
// ============================================================

func (s WaitlistStatus) String() string { return string(s) }

func (s WaitlistStatus) IsValid() bool {
	for _, status := range AllWaitlistStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func (s WaitlistStatus) GetSlug() string { return string(s) }

func (s WaitlistStatus) GetName() string {
	switch s {
	case WaitlistStatusWaiting:
		return WaitlistStatusWaitingName
	case WaitlistStatusOffered:
		return WaitlistStatusOfferedName
	case WaitlistStatusConverted:
		return WaitlistStatusConvertedName
	case WaitlistStatusExpired:
		return WaitlistStatusExpiredName
	case WaitlistStatusCancelled:
		return WaitlistStatusCancelledName
	default:
		return string(s)
	}
}

func (s WaitlistStatus) GetDisplayName() string {
	switch s {
	case WaitlistStatusWaiting:
		return WaitlistStatusWaitingDisplayName
	case WaitlistStatusOffered:
		return WaitlistStatusOfferedDisplayName
	case WaitlistStatusConverted:
		return WaitlistStatusConvertedDisplayName
	case WaitlistStatusExpired:
		return WaitlistStatusExpiredDisplayName
	case WaitlistStatusCancelled:
		return WaitlistStatusCancelledDisplayName
	default:
		return string(s)
	}
}

func (s WaitlistStatus) GetDescription() string {
	switch s {
	case WaitlistStatusWaiting:
		return WaitlistStatusWaitingDescription
	case WaitlistStatusOffered:
		return WaitlistStatusOfferedDescription
	case WaitlistStatusConverted:
		return WaitlistStatusConvertedDescription
	case WaitlistStatusExpired:
		return WaitlistStatusExpiredDescription
	case WaitlistStatusCancelled:
		return WaitlistStatusCancelledDescription
	default:
		return ""
	}
}

func (s WaitlistStatus) GetIcon() string {
	switch s {
	case WaitlistStatusWaiting:
		return WaitlistStatusWaitingIcon
	case WaitlistStatusOffered:
		return WaitlistStatusOfferedIcon
	case WaitlistStatusConverted:
		return WaitlistStatusConvertedIcon
	case WaitlistStatusExpired:
		return WaitlistStatusExpiredIcon
	case WaitlistStatusCancelled:
		return WaitlistStatusCancelledIcon
	default:
		return "circle"
	}
}

func (s WaitlistStatus) GetColor() string {
	switch s {
	case WaitlistStatusWaiting:
		return WaitlistStatusWaitingColor
	case WaitlistStatusOffered:
		return WaitlistStatusOfferedColor
	case WaitlistStatusConverted:
		return WaitlistStatusConvertedColor
	case WaitlistStatusExpired:
		return WaitlistStatusExpiredColor
	case WaitlistStatusCancelled:
		return WaitlistStatusCancelledColor
	default:
		return "#6B7280"
	}
}

func (s WaitlistStatus) GetSortOrder() int {
	switch s {
	case WaitlistStatusWaiting:
		return WaitlistStatusWaitingSortOrder
	case WaitlistStatusOffered:
		return WaitlistStatusOfferedSortOrder
	case WaitlistStatusConverted:
		return WaitlistStatusConvertedSortOrder
	case WaitlistStatusExpired:
		return WaitlistStatusExpiredSortOrder
	case WaitlistStatusCancelled:
		return WaitlistStatusCancelledSortOrder
	default:
		return 999
	}
}

func (s WaitlistStatus) IsFinal() bool {
	switch s {
	case WaitlistStatusConverted, WaitlistStatusExpired, WaitlistStatusCancelled:
		return true
	}
	return false
}

func (s WaitlistStatus) CanTransitionTo(next WaitlistStatus) bool {
	switch s {
	case WaitlistStatusWaiting:
		return next == WaitlistStatusOffered ||
			next == WaitlistStatusConverted ||
			next == WaitlistStatusCancelled
	case WaitlistStatusOffered:
		return next == WaitlistStatusConverted ||
			next == WaitlistStatusExpired ||
			next == WaitlistStatusCancelled
	case WaitlistStatusConverted, WaitlistStatusExpired, WaitlistStatusCancelled:
		return false
	}
	return false
}

// ============================================================
// PARSING
// ============================================================

func ParseWaitlistStatus(value string) (WaitlistStatus, bool) {
	s := WaitlistStatus(value)
	if s.IsValid() {
		return s, true
	}
	return "", false
}

func ParseWaitlistStatusWithDefault(value string, fallback WaitlistStatus) WaitlistStatus {
	if s, ok := ParseWaitlistStatus(value); ok {
		return s
	}
	return fallback
}