// internal/constants/attendance_statuses.go

package constants

// AttendanceStatusInfo holds information about an attendance status
type AttendanceStatusInfo struct {
	Slug                string // Database value (e.g., "registered")
	Name                string // Canonical name (e.g., "Registered")
	DisplayName         string // UI display (e.g., "📋 Registered")
	Description         string
	Color               string
	Icon                string
	SortOrder           int
	CanIssueCertificate bool
	IsFinal             bool
}

// AttendanceStatuses - the source of truth for attendance statuses
var AttendanceStatuses = []AttendanceStatusInfo{
	{
		Slug:                "registered",
		Name:                "Registered",
		DisplayName:         "📋 Registered",
		Description:         "Attendee has registered for the event but not yet joined",
		Color:               "#6B7280", // Gray
		Icon:                "clock",
		SortOrder:           1,
		CanIssueCertificate: false,
		IsFinal:             false,
	},
	{
		Slug:                "joined",
		Name:                "Joined",
		DisplayName:         "👤 Joined",
		Description:         "Attendee has joined the event",
		Color:               "#3B82F6", // Blue
		Icon:                "user-check",
		SortOrder:           2,
		CanIssueCertificate: false,
		IsFinal:             false,
	},
	{
		Slug:                "partial",
		Name:                "Partial",
		DisplayName:         "⏳ Partial",
		Description:         "Attendee joined but left early (less than full duration)",
		Color:               "#F59E0B", // Amber
		Icon:                "clock",
		SortOrder:           3,
		CanIssueCertificate: false,
		IsFinal:             true,
	},
	{
		Slug:                "full",
		Name:                "Full",
		DisplayName:         "✅ Full",
		Description:         "Attendee stayed for the full session duration",
		Color:               "#10B981", // Green
		Icon:                "check-circle",
		SortOrder:           4,
		CanIssueCertificate: true,
		IsFinal:             true,
	},
	{
		Slug:                "confirmed",
		Name:                "Confirmed",
		DisplayName:         "🟣 Confirmed",
		Description:         "Host manually confirmed attendance",
		Color:               "#8B5CF6", // Purple
		Icon:                "user-check",
		SortOrder:           5,
		CanIssueCertificate: true,
		IsFinal:             true,
	},
	{
		Slug:                "no_show",
		Name:                "No-Show",
		DisplayName:         "❌ No-Show",
		Description:         "Attendee registered but never joined",
		Color:               "#EF4444", // Red
		Icon:                "x-circle",
		SortOrder:           6,
		CanIssueCertificate: false,
		IsFinal:             true,
	},
}

// AttendanceStatusMap for quick lookups
var AttendanceStatusMap = map[string]AttendanceStatusInfo{
	"registered": AttendanceStatuses[0],
	"joined":     AttendanceStatuses[1],
	"partial":    AttendanceStatuses[2],
	"full":       AttendanceStatuses[3],
	"confirmed":  AttendanceStatuses[4],
	"no_show":    AttendanceStatuses[5],
}

// AllAttendanceStatusSlugs returns all valid attendance status slugs
func AllAttendanceStatusSlugs() []string {
	values := make([]string, len(AttendanceStatuses))
	for i, as := range AttendanceStatuses {
		values[i] = as.Slug
	}
	return values
}

// GetAttendanceStatusInfo returns AttendanceStatusInfo by slug
func GetAttendanceStatusInfo(slug string) (AttendanceStatusInfo, bool) {
	info, ok := AttendanceStatusMap[slug]
	return info, ok
}

// IsValidAttendanceStatus checks if an attendance status is valid
func IsValidAttendanceStatus(slug string) bool {
	_, ok := AttendanceStatusMap[slug]
	return ok
}

// GetStatusesThatCanIssueCertificate returns statuses that can issue certificates
func GetStatusesThatCanIssueCertificate() []AttendanceStatusInfo {
	var result []AttendanceStatusInfo
	for _, as := range AttendanceStatuses {
		if as.CanIssueCertificate {
			result = append(result, as)
		}
	}
	return result
}

// GetFinalStatuses returns final statuses
func GetFinalStatuses() []AttendanceStatusInfo {
	var result []AttendanceStatusInfo
	for _, as := range AttendanceStatuses {
		if as.IsFinal {
			result = append(result, as)
		}
	}
	return result
}