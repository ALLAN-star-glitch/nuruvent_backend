// internal/modules/registration/domain/registration_status_value.go

package registrationdomain

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"

// ============================================================
// REGISTRATION STATUS - Value Object
// ============================================================

type RegistrationStatusValue = types.RegistrationStatus

const (
	RegistrationStatusPending   = types.RegistrationStatusPending
	RegistrationStatusConfirmed = types.RegistrationStatusConfirmed
	RegistrationStatusCancelled = types.RegistrationStatusCancelled
	RegistrationStatusRefunded  = types.RegistrationStatusRefunded
	RegistrationStatusExpired   = types.RegistrationStatusExpired
	RegistrationStatusAttended  = types.RegistrationStatusAttended
)

// AllRegistrationStatuses re-exported from shared types
var AllRegistrationStatuses = types.AllRegistrationStatuses

// RegistrationStatusInfo holds metadata for each status
type RegistrationStatusInfo struct {
	Slug        string
	Name        string
	DisplayName string
	Description string
	Color       string
	Icon        string
	SortOrder   int
	IsFinal     bool
	IsActive    bool
}

// ============================================================
// REGISTRATION STATUS REGISTRY
// ============================================================

var registrationStatusRegistry = map[types.RegistrationStatus]RegistrationStatusInfo{
	types.RegistrationStatusPending: {
		Slug:        types.RegistrationStatusPendingSlug,
		Name:        types.RegistrationStatusPendingName,
		DisplayName: types.RegistrationStatusPendingDisplayName,
		Description: types.RegistrationStatusPendingDescription,
		Color:       types.RegistrationStatusPendingColor,
		Icon:        types.RegistrationStatusPendingIcon,
		SortOrder:   types.RegistrationStatusPendingSortOrder,
		IsFinal:     false,
		IsActive:    true,
	},
	types.RegistrationStatusConfirmed: {
		Slug:        types.RegistrationStatusConfirmedSlug,
		Name:        types.RegistrationStatusConfirmedName,
		DisplayName: types.RegistrationStatusConfirmedDisplayName,
		Description: types.RegistrationStatusConfirmedDescription,
		Color:       types.RegistrationStatusConfirmedColor,
		Icon:        types.RegistrationStatusConfirmedIcon,
		SortOrder:   types.RegistrationStatusConfirmedSortOrder,
		IsFinal:     false,
		IsActive:    true,
	},
	types.RegistrationStatusCancelled: {
		Slug:        types.RegistrationStatusCancelledSlug,
		Name:        types.RegistrationStatusCancelledName,
		DisplayName: types.RegistrationStatusCancelledDisplayName,
		Description: types.RegistrationStatusCancelledDescription,
		Color:       types.RegistrationStatusCancelledColor,
		Icon:        types.RegistrationStatusCancelledIcon,
		SortOrder:   types.RegistrationStatusCancelledSortOrder,
		IsFinal:     true,
		IsActive:    true,
	},
	types.RegistrationStatusRefunded: {
		Slug:        types.RegistrationStatusRefundedSlug,
		Name:        types.RegistrationStatusRefundedName,
		DisplayName: types.RegistrationStatusRefundedDisplayName,
		Description: types.RegistrationStatusRefundedDescription,
		Color:       types.RegistrationStatusRefundedColor,
		Icon:        types.RegistrationStatusRefundedIcon,
		SortOrder:   types.RegistrationStatusRefundedSortOrder,
		IsFinal:     true,
		IsActive:    true,
	},
	types.RegistrationStatusExpired: {
		Slug:        types.RegistrationStatusExpiredSlug,
		Name:        types.RegistrationStatusExpiredName,
		DisplayName: types.RegistrationStatusExpiredDisplayName,
		Description: types.RegistrationStatusExpiredDescription,
		Color:       types.RegistrationStatusExpiredColor,
		Icon:        types.RegistrationStatusExpiredIcon,
		SortOrder:   types.RegistrationStatusExpiredSortOrder,
		IsFinal:     true,
		IsActive:    true,
	},
	types.RegistrationStatusAttended: {
		Slug:        types.RegistrationStatusAttendedSlug,
		Name:        types.RegistrationStatusAttendedName,
		DisplayName: types.RegistrationStatusAttendedDisplayName,
		Description: types.RegistrationStatusAttendedDescription,
		Color:       types.RegistrationStatusAttendedColor,
		Icon:        types.RegistrationStatusAttendedIcon,
		SortOrder:   types.RegistrationStatusAttendedSortOrder,
		IsFinal:     false,
		IsActive:    true,
	},
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func GetRegistrationStatusInfo(s RegistrationStatusValue) (RegistrationStatusInfo, bool) {
	info, ok := registrationStatusRegistry[s]
	return info, ok
}

func IsRegistrationStatusValid(s RegistrationStatusValue) bool {
	return s.IsValid()
}

func AllRegistrationStatusInfos() []RegistrationStatusInfo {
	infos := make([]RegistrationStatusInfo, 0, len(registrationStatusRegistry))
	for _, info := range registrationStatusRegistry {
		infos = append(infos, info)
	}
	return infos
}

func ActiveRegistrationStatusInfos() []RegistrationStatusInfo {
	infos := make([]RegistrationStatusInfo, 0)
	for _, info := range registrationStatusRegistry {
		if info.IsActive {
			infos = append(infos, info)
		}
	}
	return infos
}

func GetRegistrationStatusBySlug(slug string) (RegistrationStatusInfo, bool) {
	for _, info := range registrationStatusRegistry {
		if info.Slug == slug {
			return info, true
		}
	}
	return RegistrationStatusInfo{}, false
}