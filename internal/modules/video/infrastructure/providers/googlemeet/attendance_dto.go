// internal/modules/video/infrastructure/providers/googlemeet/attendance_dto.go

package googlemeet

// ============================================================
// CONFERENCE RECORDS
// ============================================================

// conferenceRecordsListResponse is Google's response for
// GET /v2/conferenceRecords.
//
// Shape:
//
//	{
//	  "conferenceRecords": [ ... ],
//	  "nextPageToken": "..."
//	}
type conferenceRecordsListResponse struct {
	ConferenceRecords []conferenceRecord `json:"conferenceRecords"`
	NextPageToken     string             `json:"nextPageToken,omitempty"`
}

// conferenceRecord is a single Google conference record.
type conferenceRecord struct {
	// Name is the resource name, e.g. "conferenceRecords/abc123".
	Name string `json:"name"`

	// Space is the space resource, e.g. "spaces/efoBh0pzi5IB".
	Space string `json:"space"`

	// StartTime is when the conference began, RFC 3339.
	StartTime string `json:"startTime"`

	// EndTime is when the conference ended. Empty if still active.
	EndTime string `json:"endTime,omitempty"`

	// ExpireTime is when Google purges the record. Unused.
	ExpireTime string `json:"expireTime,omitempty"`
}

// ============================================================
// PARTICIPANTS
// ============================================================

// participantsListResponse is Google's response for
// GET /v2/{recordName}/participants.
//
// Shape:
//
//	{
//	  "participants": [ ... ],
//	  "nextPageToken": "..."
//	}
type participantsListResponse struct {
	Participants  []participant `json:"participants"`
	NextPageToken string        `json:"nextPageToken,omitempty"`
}

// participant is a single Google participant.
//
// Exactly one of SignedinUser, AnonymousUser, PhoneUser is set.
type participant struct {
	Name              string           `json:"name"`
	SignedinUser      *participantUser `json:"signedinUser,omitempty"`
	AnonymousUser     *participantUser `json:"anonymousUser,omitempty"`
	PhoneUser         *participantUser `json:"phoneUser,omitempty"`
	EarliestStartTime string           `json:"earliestStartTime"`
	LatestEndTime     string           `json:"latestEndTime,omitempty"`
}

// participantUser carries the identity of a participant. Only User
// is set for signed-in users; only DisplayName is set for
// anonymous and phone users.
type participantUser struct {
	User        string `json:"user,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}