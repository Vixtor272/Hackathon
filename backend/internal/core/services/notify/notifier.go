// Package notify fans order events out: WhatsApp to the client (through the
// MessageSender port) and the outbox for pharmacies and couriers.
package notify

import (
	"context"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Notifier implements ports.Notifier and ports.NotificationService.
type Notifier struct {
	sender ports.MessageSender
	outbox ports.NotificationRepository
	clock  ports.Clock
	ids    ports.IDGenerator
}

// New wires the channels.
func New(sender ports.MessageSender, outbox ports.NotificationRepository, clock ports.Clock, ids ports.IDGenerator) *Notifier {
	return &Notifier{sender: sender, outbox: outbox, clock: clock, ids: ids}
}

// Client sends a WhatsApp message and records it in the outbox.
func (n *Notifier) Client(ctx context.Context, phone, orderID, text string) error {
	now := n.clock.Now()
	msg := domain.Message{ID: n.ids.New("msg"), Direction: domain.DirectionOut, Type: domain.MessageText, Text: text, At: now}
	if err := n.sender.Send(ctx, phone, msg); err != nil {
		return err
	}
	return n.record(ctx, domain.ChannelWhatsApp, phone, orderID, "Mensaje de Farmi", text)
}

// Pharmacy notifies a store's cashier (simulated).
func (n *Notifier) Pharmacy(ctx context.Context, pharmacyID, orderID, title, body string) error {
	return n.record(ctx, domain.ChannelPharmacy, pharmacyID, orderID, title, body)
}

// Courier notifies the delivery person (simulated).
func (n *Notifier) Courier(ctx context.Context, courierName, orderID, title, body string) error {
	return n.record(ctx, domain.ChannelCourier, courierName, orderID, title, body)
}

// List exposes the outbox.
func (n *Notifier) List(ctx context.Context, f domain.NotificationFilter) ([]domain.Notification, error) {
	return n.outbox.List(ctx, f)
}

func (n *Notifier) record(ctx context.Context, ch domain.Channel, recipient, orderID, title, body string) error {
	return n.outbox.Save(ctx, domain.Notification{
		ID: n.ids.New("ntf"), Channel: ch, Recipient: recipient, Title: title, Body: body, OrderID: orderID, At: n.clock.Now(),
	})
}
