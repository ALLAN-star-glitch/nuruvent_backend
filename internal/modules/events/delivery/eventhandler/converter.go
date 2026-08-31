// internal/modules/events/handler/converter.go

package eventhandler

import (
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/events/service"
)

// ============================================================
// CONVERTER FUNCTIONS - JSON (No image upload)
// ============================================================

// ConvertCreateDraftRequestToCommand converts CreateDraftRequest to service.CreateDraftCommand
// ownerType and institutionID are passed from the handler (derived from URL)
func ConvertCreateDraftRequestToCommand(req CreateDraftRequest, userID, ownerType, institutionID string) service.CreateDraftCommand {
	var institutionIDPtr *string
	if institutionID != "" {
		institutionIDPtr = &institutionID
	}

	return service.CreateDraftCommand{
		Name:             req.Name,
		Description:      req.Description,
		ShortDescription: req.ShortDescription,
		EventTypeID:      req.EventTypeID,
		CategoryID:       stringPtr(req.CategoryID),
		Tags:             req.Tags,
		Language:         req.Language,
		CreatedBy:        userID,
		OwnerType:        ownerType,          // ✅ Set from handler
		InstitutionID:    institutionIDPtr,   // ✅ Set from handler

		Schedules:   convertScheduleRequestsToInputs(req.Schedules),
		IsMultiDay:  req.IsMultiDay,
		IsRecurring: req.IsRecurring,
		Recurrence:  convertRecurrenceRequestToInput(req.Recurrence),

		IsVirtual:          req.IsVirtual,
		IsHybrid:           req.IsHybrid,
		InPersonLocation:   req.InPersonLocation,
		VirtualPlatform:    req.VirtualPlatform,
		VirtualPlatformURL: req.VirtualPlatformURL,
		ZoomLink:           req.ZoomLink,
		MeetLink:           req.MeetLink,
		VenueName:          req.VenueName,
		VenueAddress:       req.VenueAddress,
		VenueCity:          req.VenueCity,
		VenueCountry:       req.VenueCountry,

		IsFree:   req.IsFree,
		Capacity: intPtr(req.Capacity),
		Tickets:  convertTicketRequestsToInputs(req.Tickets),

		Visibility:          req.Visibility,
		Password:            stringPtr(req.Password),
		InviteOnly:          req.InviteOnly,
		InvitedEmails:       req.InvitedEmails,
		IsFeatured:          req.IsFeatured,
		CertificateEnabled:  req.CertificateEnabled,
		CertificatePrice:    req.CertificatePrice,
		CertificateTemplateID: stringPtr(req.CertificateTemplateID),

		Speakers:  convertSpeakerRequestsToInputs(req.Speakers),
		Materials: convertMaterialRequestsToInputs(req.Materials),

		SEO: convertSEORequestToInput(req.SEO),
	}
}

// ConvertCreateEventRequestToCommand converts CreateEventRequest to service.CreateEventCommand
// ownerType and institutionID are passed from the handler (derived from URL)
func ConvertCreateEventRequestToCommand(req CreateEventRequest, userID, ownerType, institutionID string) service.CreateEventCommand {
	var institutionIDPtr *string
	if institutionID != "" {
		institutionIDPtr = &institutionID
	}

	return service.CreateEventCommand{
		Name:             req.Name,
		Description:      req.Description,
		ShortDescription: req.ShortDescription,
		EventTypeID:      req.EventTypeID,
		CategoryID:       stringPtr(req.CategoryID),
		Tags:             req.Tags,
		Language:         req.Language,
		CreatedBy:        userID,
		OwnerType:        ownerType,          // ✅ Set from handler
		InstitutionID:    institutionIDPtr,   // ✅ Set from handler

		Schedules:   convertScheduleRequestsToInputs(req.Schedules),
		IsMultiDay:  req.IsMultiDay,
		IsRecurring: req.IsRecurring,
		Recurrence:  convertRecurrenceRequestToInput(req.Recurrence),

		IsVirtual:          req.IsVirtual,
		IsHybrid:           req.IsHybrid,
		InPersonLocation:   req.InPersonLocation,
		VirtualPlatform:    req.VirtualPlatform,
		VirtualPlatformURL: req.VirtualPlatformURL,
		ZoomLink:           req.ZoomLink,
		MeetLink:           req.MeetLink,
		VenueName:          req.VenueName,
		VenueAddress:       req.VenueAddress,
		VenueCity:          req.VenueCity,
		VenueCountry:       req.VenueCountry,

		IsFree:   req.IsFree,
		Capacity: intPtr(req.Capacity),
		Waitlist: req.Waitlist,
		Tickets:  convertTicketRequestsToInputs(req.Tickets),

		Visibility:          req.Visibility,
		Password:            stringPtr(req.Password),
		InviteOnly:          req.InviteOnly,
		InvitedEmails:       req.InvitedEmails,
		IsFeatured:          req.IsFeatured,
		CertificateEnabled:  req.CertificateEnabled,
		CertificatePrice:    req.CertificatePrice,
		CertificateTemplateID: stringPtr(req.CertificateTemplateID),

		Speakers:  convertSpeakerRequestsToInputs(req.Speakers),
		Materials: convertMaterialRequestsToInputs(req.Materials),

		SEO: convertSEORequestToInput(req.SEO),
	}
}

// ConvertUpdateEventRequestToCommand converts UpdateEventRequest to service.UpdateEventCommand
// Note: Updates don't need owner_type or institution_id since the event already exists
func ConvertUpdateEventRequestToCommand(req UpdateEventRequest, eventID, userID string) service.UpdateEventCommand {
	return service.UpdateEventCommand{
		ID:               eventID,
		UpdatedBy:        userID,
		Name:             req.Name,
		DisplayName:      req.DisplayName,
		Description:      req.Description,
		ShortDescription: req.ShortDescription,
		EventTypeID:      req.EventTypeID,
		CategoryID:       req.CategoryID,
		Tags:             req.Tags,
		Language:         req.Language,

		Schedules:   convertScheduleRequestsToInputs(req.Schedules),
		IsMultiDay:  req.IsMultiDay,
		IsRecurring: req.IsRecurring,
		Recurrence:  convertRecurrenceRequestToInput(req.Recurrence),

		IsVirtual:          req.IsVirtual,
		IsHybrid:           req.IsHybrid,
		InPersonLocation:   req.InPersonLocation,
		VirtualPlatform:    req.VirtualPlatform,
		VirtualPlatformURL: req.VirtualPlatformURL,
		ZoomLink:           req.ZoomLink,
		MeetLink:           req.MeetLink,
		VenueName:          req.VenueName,
		VenueAddress:       req.VenueAddress,
		VenueCity:          req.VenueCity,
		VenueCountry:       req.VenueCountry,

		IsFree:   req.IsFree,
		Capacity: req.Capacity,
		Waitlist: req.Waitlist,
		Tickets:  convertTicketRequestsToInputs(req.Tickets),

		Visibility:          req.Visibility,
		Password:            req.Password,
		InviteOnly:          req.InviteOnly,
		InvitedEmails:       req.InvitedEmails,
		IsFeatured:          req.IsFeatured,
		CertificateEnabled:  req.CertificateEnabled,
		CertificatePrice:    req.CertificatePrice,
		CertificateTemplateID: req.CertificateTemplateID,

		Speakers:  convertSpeakerRequestsToInputs(req.Speakers),
		Materials: convertMaterialRequestsToInputs(req.Materials),

		SEO: convertSEORequestToInput(req.SEO),
	}
}

// ============================================================
// CONVERTER FUNCTIONS - Request Types to Input Types
// ============================================================

func convertScheduleRequestsToInputs(reqs []ScheduleRequest) []service.ScheduleInput {
	if len(reqs) == 0 {
		return nil
	}
	inputs := make([]service.ScheduleInput, len(reqs))
	for i, r := range reqs {
		inputs[i] = service.ScheduleInput{
			ID:            stringPtr(r.ID),
			StartDate:     r.StartDate,
			EndDate:       stringPtr(r.EndDate),
			StartTime:     r.StartTime,
			EndTime:       r.EndTime,
			Timezone:      r.Timezone,
			SessionName:   r.SessionName,
			SessionNumber: r.SessionNumber,
			Location:      r.Location,
			IsVirtual:     r.IsVirtual,
			ZoomLink:      r.ZoomLink,
			MeetLink:      r.MeetLink,
			MaxAttendees:  r.MaxAttendees,
		}
	}
	return inputs
}

func convertTicketRequestsToInputs(reqs []TicketRequest) []service.TicketInput {
	if len(reqs) == 0 {
		return nil
	}
	inputs := make([]service.TicketInput, len(reqs))
	for i, r := range reqs {
		inputs[i] = service.TicketInput{
			ID:                 stringPtr(r.ID),
			TicketTypeID:       r.TicketTypeID,
			Name:               r.Name,
			Description:        r.Description,
			Price:              r.Price,
			Quantity:           r.Quantity,
			MaxPerPerson:       r.MaxPerPerson,
			EarlyBirdDeadline:  stringPtr(r.EarlyBirdDeadline),
			GroupMinAttendees:  r.GroupMinAttendees,
			GroupDiscount:      r.GroupDiscount,
		}
	}
	return inputs
}

func convertSpeakerRequestsToInputs(reqs []SpeakerRequest) []service.SpeakerInput {
	if len(reqs) == 0 {
		return nil
	}
	inputs := make([]service.SpeakerInput, len(reqs))
	for i, r := range reqs {
		inputs[i] = service.SpeakerInput{
			ID:          stringPtr(r.ID),
			Name:        r.Name,
			Title:       r.Title,
			Bio:         r.Bio,
			PhotoURL:    r.PhotoURL,
			SocialLinks: r.SocialLinks,
			IsKeynote:   r.IsKeynote,
			SortOrder:   r.SortOrder,
		}
	}
	return inputs
}

func convertMaterialRequestsToInputs(reqs []MaterialRequest) []service.MaterialInput {
	if len(reqs) == 0 {
		return nil
	}
	inputs := make([]service.MaterialInput, len(reqs))
	for i, r := range reqs {
		inputs[i] = service.MaterialInput{
			ID:             stringPtr(r.ID),
			Title:          r.Title,
			MaterialTypeID: r.MaterialTypeID,
			URL:            r.URL,
			Description:    r.Description,
			IsPreEvent:     r.IsPreEvent,
			SortOrder:      r.SortOrder,
		}
	}
	return inputs
}

func convertRecurrenceRequestToInput(req *RecurrenceRequest) *service.RecurrenceInput {
	if req == nil {
		return nil
	}
	return &service.RecurrenceInput{
		Pattern:     req.Pattern,
		Interval:    req.Interval,
		DaysOfWeek:  req.DaysOfWeek,
		DayOfMonth:  req.DayOfMonth,
		WeekOfMonth: stringPtr(req.WeekOfMonth),
		EndsOn:      stringPtr(req.EndsOn),
		Occurrences: req.Occurrences,
	}
}

func convertSEORequestToInput(req *SEORequest) *service.SEOInput {
	if req == nil {
		return nil
	}
	return &service.SEOInput{
		MetaTitle:          req.MetaTitle,
		MetaDescription:    req.MetaDescription,
		MetaKeywords:       req.MetaKeywords,
		CanonicalURL:       req.CanonicalURL,
		Robots:             req.Robots,
		NoIndex:            req.NoIndex,
		OGTitle:            req.OGTitle,
		OGDescription:      req.OGDescription,
		OGImageURL:         req.OGImageURL,
		OGType:             req.OGType,
		TwitterCard:        req.TwitterCard,
		TwitterTitle:       req.TwitterTitle,
		TwitterDescription: req.TwitterDescription,
		TwitterImageURL:    req.TwitterImageURL,
	}
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}