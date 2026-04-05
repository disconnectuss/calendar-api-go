package webhooks

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Service struct {
	mu       sync.RWMutex
	channels map[string]*WebhookChannel
}

func NewService() *Service {
	return &Service{
		channels: make(map[string]*WebhookChannel),
	}
}

func (s *Service) Subscribe(ctx context.Context, req *CreateWebhookRequest, opts ...option.ClientOption) (*WebhookChannel, error) {
	svc, err := gcal.NewService(ctx, opts...)
	if err != nil {
		return nil, err
	}

	channelID := uuid.New().String()
	expiration := time.Now().Add(7 * 24 * time.Hour) // 7 days

	channel := &gcal.Channel{
		Id:         channelID,
		Type:       "web_hook",
		Address:    req.WebhookURL,
		Expiration: expiration.UnixMilli(),
	}

	result, err := svc.Events.Watch(req.CalendarID, channel).Do()
	if err != nil {
		return nil, err
	}

	wc := &WebhookChannel{
		ID:         result.Id,
		ResourceID: result.ResourceId,
		CalendarID: req.CalendarID,
		Expiration: expiration.Format(time.RFC3339),
	}

	s.mu.Lock()
	s.channels[result.Id] = wc
	s.mu.Unlock()

	return wc, nil
}

func (s *Service) Unsubscribe(ctx context.Context, req *StopWebhookRequest, opts ...option.ClientOption) error {
	svc, err := gcal.NewService(ctx, opts...)
	if err != nil {
		return err
	}

	channel := &gcal.Channel{
		Id:         req.ChannelID,
		ResourceId: req.ResourceID,
	}

	if err := svc.Channels.Stop(channel).Do(); err != nil {
		return err
	}

	s.mu.Lock()
	delete(s.channels, req.ChannelID)
	s.mu.Unlock()

	return nil
}

func (s *Service) ListChannels() []*WebhookChannel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	channels := make([]*WebhookChannel, 0, len(s.channels))
	for _, ch := range s.channels {
		channels = append(channels, ch)
	}
	return channels
}
