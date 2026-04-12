package calendar

import (
	"context"
	"fmt"
	"time"

	"api-go/internal/common"

	"github.com/patrickmn/go-cache"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Service struct {
	cache *cache.Cache
}

func NewService() *Service {
	return &Service{
		cache: cache.New(30*time.Second, 60*time.Second),
	}
}

func (s *Service) getCalendarService(ctx context.Context, opts ...option.ClientOption) (*gcal.Service, error) {
	return gcal.NewService(ctx, opts...)
}

func (s *Service) ListEvents(ctx context.Context, maxResults int64, pageToken string, opts ...option.ClientOption) (*common.PaginatedResponse[*gcal.Event], error) {
	cacheKey := fmt.Sprintf("calendar:events:%d:%s", maxResults, pageToken)
	if cached, found := s.cache.Get(cacheKey); found {
		result := cached.(*common.PaginatedResponse[*gcal.Event])
		return result, nil
	}

	svc, err := s.getCalendarService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	call := svc.Events.List("primary").MaxResults(maxResults).SingleEvents(true).OrderBy("startTime")
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	events, err := call.Do()
	if err != nil {
		return nil, err
	}

	result := common.NewPaginatedResponse(events.Items, events.NextPageToken)
	s.cache.Set(cacheKey, &result, cache.DefaultExpiration)
	return &result, nil
}

func (s *Service) GetEvent(ctx context.Context, eventID string, opts ...option.ClientOption) (*gcal.Event, error) {
	svc, err := s.getCalendarService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	event, err := svc.Events.Get("primary", eventID).Do()
	if err != nil {
		return nil, err
	}
	return event, nil
}

func (s *Service) CreateGeneralEvent(ctx context.Context, req *CreateGeneralEventRequest, opts ...option.ClientOption) (*gcal.Event, error) {
	svc, err := s.getCalendarService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	event := buildEvent(req)
	event.Visibility = "public"

	created, err := svc.Events.Insert("primary", event).Do()
	if err != nil {
		return nil, err
	}

	s.invalidateCache()
	return created, nil
}

func (s *Service) CreatePrivateEvent(ctx context.Context, req *CreatePrivateEventRequest, opts ...option.ClientOption) (*gcal.Event, error) {
	svc, err := s.getCalendarService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	event := buildEvent(&req.CreateGeneralEventRequest)
	event.Visibility = "private"

	for _, email := range req.AllowedUsers {
		event.Attendees = append(event.Attendees, &gcal.EventAttendee{
			Email: email,
		})
	}

	created, err := svc.Events.Insert("primary", event).Do()
	if err != nil {
		return nil, err
	}

	s.invalidateCache()
	return created, nil
}

func (s *Service) UpdateEvent(ctx context.Context, eventID string, req *UpdateEventRequest, userEmail string, opts ...option.ClientOption) (*gcal.Event, error) {
	svc, err := s.getCalendarService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	existing, err := svc.Events.Get("primary", eventID).Do()
	if err != nil {
		return nil, err
	}

	if !isOrganizerOrAttendee(existing, userEmail) {
		return nil, common.ForbiddenError("You don't have permission to update this event")
	}

	if req.Summary != "" {
		existing.Summary = req.Summary
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Start != nil {
		existing.Start = &gcal.EventDateTime{
			DateTime: req.Start.DateTime,
			TimeZone: req.Start.TimeZone,
		}
	}
	if req.End != nil {
		existing.End = &gcal.EventDateTime{
			DateTime: req.End.DateTime,
			TimeZone: req.End.TimeZone,
		}
	}
	if req.Location != "" {
		existing.Location = req.Location
	}
	if req.Attendees != nil {
		existing.Attendees = buildAttendees(req.Attendees)
	}

	updated, err := svc.Events.Patch("primary", eventID, existing).Do()
	if err != nil {
		return nil, err
	}

	s.invalidateCache()
	return updated, nil
}

func (s *Service) DeleteEvent(ctx context.Context, eventID string, userEmail string, opts ...option.ClientOption) error {
	svc, err := s.getCalendarService(ctx, opts...)
	if err != nil {
		return err
	}

	existing, err := svc.Events.Get("primary", eventID).Do()
	if err != nil {
		return err
	}

	if !isOrganizer(existing, userEmail) {
		return common.ForbiddenError("Only the organizer can delete this event")
	}

	if err := svc.Events.Delete("primary", eventID).Do(); err != nil {
		return err
	}

	s.invalidateCache()
	return nil
}

func (s *Service) invalidateCache() {
	s.cache.Flush()
}

func buildEvent(req *CreateGeneralEventRequest) *gcal.Event {
	event := &gcal.Event{
		Summary:     req.Summary,
		Description: req.Description,
		Location:    req.Location,
		Start: &gcal.EventDateTime{
			DateTime: req.Start.DateTime,
			TimeZone: req.Start.TimeZone,
		},
		End: &gcal.EventDateTime{
			DateTime: req.End.DateTime,
			TimeZone: req.End.TimeZone,
		},
		AnyoneCanAddSelf:        req.AnyoneCanAddSelf,
		GuestsCanInviteOthers:   &req.GuestsCanInviteOthers,
		GuestsCanModify:         req.GuestsCanModify,
		GuestsCanSeeOtherGuests: &req.GuestsCanSeeOthers,
		Attendees:               buildAttendees(req.Attendees),
	}
	return event
}

func buildAttendees(attendees []AttendeeDTO) []*gcal.EventAttendee {
	if len(attendees) == 0 {
		return nil
	}
	result := make([]*gcal.EventAttendee, len(attendees))
	for i, a := range attendees {
		result[i] = &gcal.EventAttendee{
			Email:          a.Email,
			DisplayName:    a.DisplayName,
			Optional:       a.Optional,
			ResponseStatus: a.ResponseStatus,
		}
	}
	return result
}

func isOrganizer(event *gcal.Event, email string) bool {
	if event.Organizer != nil && event.Organizer.Email == email {
		return true
	}
	return false
}

func isOrganizerOrAttendee(event *gcal.Event, email string) bool {
	if isOrganizer(event, email) {
		return true
	}
	for _, a := range event.Attendees {
		if a.Email == email {
			return true
		}
	}
	return false
}
