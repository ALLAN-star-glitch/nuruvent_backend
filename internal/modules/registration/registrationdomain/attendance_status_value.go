// internal/modules/registration/registrationdomain/attendance_status_value.go

package registrationdomain

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"

// ============================================================
// ATTENDANCE STATUS - Value Object
// ============================================================

type AttendanceStatusValue = types.AttendanceStatus

const (
	AttendanceStatusRegistered = types.AttendanceStatusRegistered
	AttendanceStatusJoined     = types.AttendanceStatusJoined
	AttendanceStatusPartial    = types.AttendanceStatusPartial
	AttendanceStatusFull       = types.AttendanceStatusFull
	AttendanceStatusConfirmed  = types.AttendanceStatusConfirmed
	AttendanceStatusNoShow     = types.AttendanceStatusNoShow
)

var AllAttendanceStatuses = types.AllAttendanceStatuses

// AttendanceStatusInfo holds metadata for each attendance status
type AttendanceStatusInfo struct {
	Slug             string
	Name             string
	DisplayName      string
	Description      string
	Color            string
	Icon             string
	SortOrder        int
	IsFinal          bool
	IsActive         bool
	CountsAsAttended bool
}

// ============================================================
// ATTENDANCE STATUS REGISTRY
// ============================================================

var attendanceStatusRegistry = map[types.AttendanceStatus]AttendanceStatusInfo{
	types.AttendanceStatusRegistered: {
		Slug:             types.AttendanceStatusRegisteredSlug,
		Name:             types.AttendanceStatusRegisteredName,
		DisplayName:      types.AttendanceStatusRegisteredDisplayName,
		Description:      types.AttendanceStatusRegisteredDescription,
		Color:            types.AttendanceStatusRegisteredColor,
		Icon:             types.AttendanceStatusRegisteredIcon,
		SortOrder:        types.AttendanceStatusRegisteredSortOrder,
		IsFinal:          false,
		IsActive:         true,
		CountsAsAttended: false,
	},
	types.AttendanceStatusJoined: {
		Slug:             types.AttendanceStatusJoinedSlug,
		Name:             types.AttendanceStatusJoinedName,
		DisplayName:      types.AttendanceStatusJoinedDisplayName,
		Description:      types.AttendanceStatusJoinedDescription,
		Color:            types.AttendanceStatusJoinedColor,
		Icon:             types.AttendanceStatusJoinedIcon,
		SortOrder:        types.AttendanceStatusJoinedSortOrder,
		IsFinal:          false,
		IsActive:         true,
		CountsAsAttended: false,
	},
	types.AttendanceStatusPartial: {
		Slug:             types.AttendanceStatusPartialSlug,
		Name:             types.AttendanceStatusPartialName,
		DisplayName:      types.AttendanceStatusPartialDisplayName,
		Description:      types.AttendanceStatusPartialDescription,
		Color:            types.AttendanceStatusPartialColor,
		Icon:             types.AttendanceStatusPartialIcon,
		SortOrder:        types.AttendanceStatusPartialSortOrder,
		IsFinal:          false,
		IsActive:         true,
		CountsAsAttended: false,
	},
	types.AttendanceStatusFull: {
		Slug:             types.AttendanceStatusFullSlug,
		Name:             types.AttendanceStatusFullName,
		DisplayName:      types.AttendanceStatusFullDisplayName,
		Description:      types.AttendanceStatusFullDescription,
		Color:            types.AttendanceStatusFullColor,
		Icon:             types.AttendanceStatusFullIcon,
		SortOrder:        types.AttendanceStatusFullSortOrder,
		IsFinal:          true,
		IsActive:         true,
		CountsAsAttended: true,
	},
	types.AttendanceStatusConfirmed: {
		Slug:             types.AttendanceStatusConfirmedSlug,
		Name:             types.AttendanceStatusConfirmedName,
		DisplayName:      types.AttendanceStatusConfirmedDisplayName,
		Description:      types.AttendanceStatusConfirmedDescription,
		Color:            types.AttendanceStatusConfirmedColor,
		Icon:             types.AttendanceStatusConfirmedIcon,
		SortOrder:        types.AttendanceStatusConfirmedSortOrder,
		IsFinal:          true,
		IsActive:         true,
		CountsAsAttended: true,
	},
	types.AttendanceStatusNoShow: {
		Slug:             types.AttendanceStatusNoShowSlug,
		Name:             types.AttendanceStatusNoShowName,
		DisplayName:      types.AttendanceStatusNoShowDisplayName,
		Description:      types.AttendanceStatusNoShowDescription,
		Color:            types.AttendanceStatusNoShowColor,
		Icon:             types.AttendanceStatusNoShowIcon,
		SortOrder:        types.AttendanceStatusNoShowSortOrder,
		IsFinal:          true,
		IsActive:         false,
		CountsAsAttended: false,
	},
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func GetAttendanceStatusInfo(s AttendanceStatusValue) (AttendanceStatusInfo, bool) {
	info, ok := attendanceStatusRegistry[s]
	return info, ok
}

func IsAttendanceStatusValid(s AttendanceStatusValue) bool {
	return s.IsValid()
}

func AllAttendanceStatusInfos() []AttendanceStatusInfo {
	infos := make([]AttendanceStatusInfo, 0, len(attendanceStatusRegistry))
	for _, info := range attendanceStatusRegistry {
		infos = append(infos, info)
	}
	return infos
}

func ActiveAttendanceStatusInfos() []AttendanceStatusInfo {
	infos := make([]AttendanceStatusInfo, 0)
	for _, info := range attendanceStatusRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

func GetAttendanceStatusBySlug(slug string) (AttendanceStatusInfo, bool) {
	for _, info := range attendanceStatusRegistry {
		if info.Slug == slug {
			return info, true
		}
	}
	return AttendanceStatusInfo{}, false
}