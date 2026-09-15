// internal/modules/registration/registrationdomain/waitlist_status_value.go

package registrationdomain

import "github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/types"

type WaitlistStatusValue = types.WaitlistStatus


var AllWaitlistStatuses = types.AllWaitlistStatuses

type WaitlistStatusInfo struct {
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

var waitlistStatusRegistry = map[types.WaitlistStatus]WaitlistStatusInfo{
	types.WaitlistStatusWaiting: {
		Slug:        types.WaitlistStatusWaitingSlug,
		Name:        types.WaitlistStatusWaitingName,
		DisplayName: types.WaitlistStatusWaitingDisplayName,
		Description: types.WaitlistStatusWaitingDescription,
		Color:       types.WaitlistStatusWaitingColor,
		Icon:        types.WaitlistStatusWaitingIcon,
		SortOrder:   types.WaitlistStatusWaitingSortOrder,
		IsFinal:     false,
		IsActive:    true,
	},
	types.WaitlistStatusOffered: {
		Slug:        types.WaitlistStatusOfferedSlug,
		Name:        types.WaitlistStatusOfferedName,
		DisplayName: types.WaitlistStatusOfferedDisplayName,
		Description: types.WaitlistStatusOfferedDescription,
		Color:       types.WaitlistStatusOfferedColor,
		Icon:        types.WaitlistStatusOfferedIcon,
		SortOrder:   types.WaitlistStatusOfferedSortOrder,
		IsFinal:     false,
		IsActive:    true,
	},
	types.WaitlistStatusConverted: {
		Slug:        types.WaitlistStatusConvertedSlug,
		Name:        types.WaitlistStatusConvertedName,
		DisplayName: types.WaitlistStatusConvertedDisplayName,
		Description: types.WaitlistStatusConvertedDescription,
		Color:       types.WaitlistStatusConvertedColor,
		Icon:        types.WaitlistStatusConvertedIcon,
		SortOrder:   types.WaitlistStatusConvertedSortOrder,
		IsFinal:     true,
		IsActive:    false,
	},
	types.WaitlistStatusExpired: {
		Slug:        types.WaitlistStatusExpiredSlug,
		Name:        types.WaitlistStatusExpiredName,
		DisplayName: types.WaitlistStatusExpiredDisplayName,
		Description: types.WaitlistStatusExpiredDescription,
		Color:       types.WaitlistStatusExpiredColor,
		Icon:        types.WaitlistStatusExpiredIcon,
		SortOrder:   types.WaitlistStatusExpiredSortOrder,
		IsFinal:     true,
		IsActive:    false,
	},
	types.WaitlistStatusCancelled: {
		Slug:        types.WaitlistStatusCancelledSlug,
		Name:        types.WaitlistStatusCancelledName,
		DisplayName: types.WaitlistStatusCancelledDisplayName,
		Description: types.WaitlistStatusCancelledDescription,
		Color:       types.WaitlistStatusCancelledColor,
		Icon:        types.WaitlistStatusCancelledIcon,
		SortOrder:   types.WaitlistStatusCancelledSortOrder,
		IsFinal:     true,
		IsActive:    false,
	},
}

func GetWaitlistStatusInfo(s WaitlistStatusValue) (WaitlistStatusInfo, bool) {
	info, ok := waitlistStatusRegistry[s]
	return info, ok
}

func AllWaitlistStatusInfos() []WaitlistStatusInfo {
	infos := make([]WaitlistStatusInfo, 0, len(waitlistStatusRegistry))
	for _, info := range waitlistStatusRegistry {
		infos = append(infos, info)
	}
	return infos
}

func GetWaitlistStatusBySlug(slug string) (WaitlistStatusInfo, bool) {
	for _, info := range waitlistStatusRegistry {
		if info.Slug == slug {
			return info, true
		}
	}
	return WaitlistStatusInfo{}, false
}