package calendar

type DateTimeDTO struct {
	DateTime string `json:"dateTime" validate:"required"`
	TimeZone string `json:"timeZone,omitempty"`
}

type AttendeeDTO struct {
	Email          string `json:"email" validate:"required,email"`
	DisplayName    string `json:"displayName,omitempty"`
	Optional       bool   `json:"optional,omitempty"`
	ResponseStatus string `json:"responseStatus,omitempty"`
}

type CreateGeneralEventRequest struct {
	Summary               string        `json:"summary" validate:"required"`
	Description           string        `json:"description,omitempty"`
	Start                 DateTimeDTO   `json:"start" validate:"required"`
	End                   DateTimeDTO   `json:"end" validate:"required"`
	Location              string        `json:"location,omitempty"`
	Attendees             []AttendeeDTO `json:"attendees,omitempty"`
	AnyoneCanAddSelf      bool          `json:"anyoneCanAddSelf,omitempty"`
	GuestsCanInviteOthers bool          `json:"guestsCanInviteOthers,omitempty"`
	GuestsCanModify       bool          `json:"guestsCanModify,omitempty"`
	GuestsCanSeeOthers    bool          `json:"guestsCanSeeOtherGuests,omitempty"`
}

type CreatePrivateEventRequest struct {
	CreateGeneralEventRequest
	AllowedUsers []string `json:"allowedUsers" validate:"required,dive,email"`
}

type UpdateEventRequest struct {
	Summary     string        `json:"summary,omitempty"`
	Description string        `json:"description,omitempty"`
	Start       *DateTimeDTO  `json:"start,omitempty"`
	End         *DateTimeDTO  `json:"end,omitempty"`
	Location    string        `json:"location,omitempty"`
	Attendees   []AttendeeDTO `json:"attendees,omitempty"`
}
