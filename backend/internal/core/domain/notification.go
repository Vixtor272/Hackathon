package domain

import "time"

// Channel is a simulated outbound channel.
type Channel string

const (
	ChannelWhatsApp Channel = "whatsapp"
	ChannelPharmacy Channel = "pharmacy"
	ChannelCourier  Channel = "courier"
)

// Notification is an entry of the outbox the demo UI can inspect.
type Notification struct {
	ID        string
	Channel   Channel
	Recipient string
	Title     string
	Body      string
	OrderID   string
	At        time.Time
}

// NotificationFilter narrows an outbox query; empty fields match everything.
type NotificationFilter struct {
	Channel Channel
	OrderID string
}
