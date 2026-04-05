package webhooks

type CreateWebhookRequest struct {
	CalendarID string `json:"calendarId" validate:"required"`
	WebhookURL string `json:"webhookUrl" validate:"required,url"`
}

type StopWebhookRequest struct {
	ChannelID  string `json:"channelId" validate:"required"`
	ResourceID string `json:"resourceId" validate:"required"`
}

type WebhookChannel struct {
	ID         string `json:"id"`
	ResourceID string `json:"resourceId"`
	CalendarID string `json:"calendarId"`
	Expiration string `json:"expiration"`
}
