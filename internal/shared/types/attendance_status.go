// internal/shared/types/attendance_status.go

package types

// ============================================================
// ATTENDANCE STATUS CONSTANTS
// ============================================================

// AttendanceStatusSlug constants (with hyphens) - Used for URLs and API routes
const (
	AttendanceStatusRegisteredSlug = "registered"
	AttendanceStatusJoinedSlug     = "joined"
	AttendanceStatusPartialSlug    = "partial"
	AttendanceStatusFullSlug       = "full"
	AttendanceStatusConfirmedSlug  = "confirmed"
	AttendanceStatusNoShowSlug     = "no-show"
)

// AttendanceStatusName constants (with underscores) - Used for internal database lookups
const (
	AttendanceStatusRegisteredName = "attendance_status_registered"
	AttendanceStatusJoinedName     = "attendance_status_joined"
	AttendanceStatusPartialName    = "attendance_status_partial"
	AttendanceStatusFullName       = "attendance_status_full"
	AttendanceStatusConfirmedName  = "attendance_status_confirmed"
	AttendanceStatusNoShowName     = "attendance_status_no_show"
)

// AttendanceStatusDisplayName constants - Used for UI display
const (
	AttendanceStatusRegisteredDisplayName = "Registered"
	AttendanceStatusJoinedDisplayName     = "Joined"
	AttendanceStatusPartialDisplayName    = "Partial"
	AttendanceStatusFullDisplayName       = "Full"
	AttendanceStatusConfirmedDisplayName  = "Confirmed"
	AttendanceStatusNoShowDisplayName     = "No Show"
)

// AttendanceStatusDescription constants
const (
	AttendanceStatusRegisteredDescription = "Registered for event but not yet joined"
	AttendanceStatusJoinedDescription     = "Joined the session but attendance not confirmed"
	AttendanceStatusPartialDescription    = "Joined but left before the session ended"
	AttendanceStatusFullDescription       = "Stayed for the full session duration"
	AttendanceStatusConfirmedDescription  = "Manually confirmed as attended by the host"
	AttendanceStatusNoShowDescription     = "Registered but did not attend"
)

// AttendanceStatusIcon constants
const (
	AttendanceStatusRegisteredIcon = "clipboard-list"
	AttendanceStatusJoinedIcon     = "log-in"
	AttendanceStatusPartialIcon    = "clock"
	AttendanceStatusFullIcon       = "check-circle"
	AttendanceStatusConfirmedIcon  = "user-check"
	AttendanceStatusNoShowIcon     = "x-circle"
)

// AttendanceStatusColor constants
const (
	AttendanceStatusRegisteredColor = "#6B7280" // Gray
	AttendanceStatusJoinedColor     = "#3B82F6" // Blue
	AttendanceStatusPartialColor    = "#F59E0B" // Amber
	AttendanceStatusFullColor       = "#10B981" // Green
	AttendanceStatusConfirmedColor  = "#059669" // Dark Green
	AttendanceStatusNoShowColor     = "#EF4444" // Red
)

// AttendanceStatusSortOrder constants
const (
	AttendanceStatusRegisteredSortOrder = 1
	AttendanceStatusJoinedSortOrder     = 2
	AttendanceStatusPartialSortOrder    = 3
	AttendanceStatusFullSortOrder       = 4
	AttendanceStatusConfirmedSortOrder  = 5
	AttendanceStatusNoShowSortOrder     = 6
)

// AttendanceStatusIsFinal constants
const (
	AttendanceStatusRegisteredIsFinal = false
	AttendanceStatusJoinedIsFinal     = false
	AttendanceStatusPartialIsFinal    = false
	AttendanceStatusFullIsFinal       = true
	AttendanceStatusConfirmedIsFinal  = true
	AttendanceStatusNoShowIsFinal     = true
)

// ============================================================
// ATTENDANCE STATUS DEFINITIONS
// ============================================================

// AttendanceStatus represents the attendance state of a registration
type AttendanceStatus string

// AttendanceStatus constants - These are the NAMES (with underscores)
const (
	AttendanceStatusRegistered AttendanceStatus = AttendanceStatusRegisteredName
	AttendanceStatusJoined     AttendanceStatus = AttendanceStatusJoinedName
	AttendanceStatusPartial    AttendanceStatus = AttendanceStatusPartialName
	AttendanceStatusFull       AttendanceStatus = AttendanceStatusFullName
	AttendanceStatusConfirmed  AttendanceStatus = AttendanceStatusConfirmedName
	AttendanceStatusNoShow     AttendanceStatus = AttendanceStatusNoShowName
)

// AllAttendanceStatuses lists all valid statuses for validation
var AllAttendanceStatuses = []AttendanceStatus{
	AttendanceStatusRegistered,
	AttendanceStatusJoined,
	AttendanceStatusPartial,
	AttendanceStatusFull,
	AttendanceStatusConfirmed,
	AttendanceStatusNoShow,
}

// ============================================================
// BASIC METHODS - On AttendanceStatus
// ============================================================

func (s AttendanceStatus) String() string {
	return string(s)
}

func (s AttendanceStatus) IsValid() bool {
	for _, status := range AllAttendanceStatuses {
		if status == s {
			return true
		}
	}
	return false
}

// ============================================================
// GETTER METHODS - On AttendanceStatus
// ============================================================

func (s AttendanceStatus) GetName() string {
	return string(s)
}

func (s AttendanceStatus) GetSlug() string {
	switch s {
	case AttendanceStatusRegistered:
		return AttendanceStatusRegisteredSlug
	case AttendanceStatusJoined:
		return AttendanceStatusJoinedSlug
	case AttendanceStatusPartial:
		return AttendanceStatusPartialSlug
	case AttendanceStatusFull:
		return AttendanceStatusFullSlug
	case AttendanceStatusConfirmed:
		return AttendanceStatusConfirmedSlug
	case AttendanceStatusNoShow:
		return AttendanceStatusNoShowSlug
	default:
		return string(s)
	}
}

func (s AttendanceStatus) GetDisplayName() string {
	switch s {
	case AttendanceStatusRegistered:
		return AttendanceStatusRegisteredDisplayName
	case AttendanceStatusJoined:
		return AttendanceStatusJoinedDisplayName
	case AttendanceStatusPartial:
		return AttendanceStatusPartialDisplayName
	case AttendanceStatusFull:
		return AttendanceStatusFullDisplayName
	case AttendanceStatusConfirmed:
		return AttendanceStatusConfirmedDisplayName
	case AttendanceStatusNoShow:
		return AttendanceStatusNoShowDisplayName
	default:
		return string(s)
	}
}

func (s AttendanceStatus) GetDescription() string {
	switch s {
	case AttendanceStatusRegistered:
		return AttendanceStatusRegisteredDescription
	case AttendanceStatusJoined:
		return AttendanceStatusJoinedDescription
	case AttendanceStatusPartial:
		return AttendanceStatusPartialDescription
	case AttendanceStatusFull:
		return AttendanceStatusFullDescription
	case AttendanceStatusConfirmed:
		return AttendanceStatusConfirmedDescription
	case AttendanceStatusNoShow:
		return AttendanceStatusNoShowDescription
	default:
		return ""
	}
}

func (s AttendanceStatus) GetIcon() string {
	switch s {
	case AttendanceStatusRegistered:
		return AttendanceStatusRegisteredIcon
	case AttendanceStatusJoined:
		return AttendanceStatusJoinedIcon
	case AttendanceStatusPartial:
		return AttendanceStatusPartialIcon
	case AttendanceStatusFull:
		return AttendanceStatusFullIcon
	case AttendanceStatusConfirmed:
		return AttendanceStatusConfirmedIcon
	case AttendanceStatusNoShow:
		return AttendanceStatusNoShowIcon
	default:
		return "circle"
	}
}

func (s AttendanceStatus) GetColor() string {
	switch s {
	case AttendanceStatusRegistered:
		return AttendanceStatusRegisteredColor
	case AttendanceStatusJoined:
		return AttendanceStatusJoinedColor
	case AttendanceStatusPartial:
		return AttendanceStatusPartialColor
	case AttendanceStatusFull:
		return AttendanceStatusFullColor
	case AttendanceStatusConfirmed:
		return AttendanceStatusConfirmedColor
	case AttendanceStatusNoShow:
		return AttendanceStatusNoShowColor
	default:
		return "#6B7280"
	}
}

func (s AttendanceStatus) GetSortOrder() int {
	switch s {
	case AttendanceStatusRegistered:
		return AttendanceStatusRegisteredSortOrder
	case AttendanceStatusJoined:
		return AttendanceStatusJoinedSortOrder
	case AttendanceStatusPartial:
		return AttendanceStatusPartialSortOrder
	case AttendanceStatusFull:
		return AttendanceStatusFullSortOrder
	case AttendanceStatusConfirmed:
		return AttendanceStatusConfirmedSortOrder
	case AttendanceStatusNoShow:
		return AttendanceStatusNoShowSortOrder
	default:
		return 999
	}
}

// IsFinal returns whether this status is terminal.
func (s AttendanceStatus) IsFinal() bool {
	switch s {
	case AttendanceStatusFull, AttendanceStatusConfirmed, AttendanceStatusNoShow:
		return true
	default:
		return false
	}
}

// CountsAsAttended returns whether this status qualifies for a certificate.
// Per BRD-ATT-10: only Full and Confirmed qualify.
func (s AttendanceStatus) CountsAsAttended() bool {
	switch s {
	case AttendanceStatusFull, AttendanceStatusConfirmed:
		return true
	default:
		return false
	}
}

// ============================================================
// PARSE FUNCTIONS
// ============================================================

func ParseAttendanceStatus(name string) (AttendanceStatus, bool) {
	s := AttendanceStatus(name)
	if s.IsValid() {
		return s, true
	}
	return "", false
}

func ParseAttendanceStatusWithDefault(name string, defaultStatus AttendanceStatus) AttendanceStatus {
	if s, ok := ParseAttendanceStatus(name); ok {
		return s
	}
	return defaultStatus
}

func ParseAttendanceStatusBySlug(slug string) (AttendanceStatus, bool) {
	for _, s := range AllAttendanceStatuses {
		if s.GetSlug() == slug {
			return s, true
		}
	}
	return "", false
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func AllAttendanceStatusNames() []string {
	names := make([]string, 0, len(AllAttendanceStatuses))
	for _, s := range AllAttendanceStatuses {
		names = append(names, s.GetName())
	}
	return names
}

func AllAttendanceStatusSlugs() []string {
	slugs := make([]string, 0, len(AllAttendanceStatuses))
	for _, s := range AllAttendanceStatuses {
		slugs = append(slugs, s.GetSlug())
	}
	return slugs
}

func AllAttendanceStatusDisplayNames() []string {
	names := make([]string, 0, len(AllAttendanceStatuses))
	for _, s := range AllAttendanceStatuses {
		names = append(names, s.GetDisplayName())
	}
	return names
}

func IsTerminalAttendanceStatus(name string) bool {
	s, ok := ParseAttendanceStatus(name)
	if !ok {
		return false
	}
	return s.IsFinal()
}

func CountsAsAttended(name string) bool {
	s, ok := ParseAttendanceStatus(name)
	if !ok {
		return false
	}
	return s.CountsAsAttended()
}