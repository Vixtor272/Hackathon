// Package memory holds every repository of the demo in process memory. Each
// store implements one port, so a SQL implementation can replace any of them
// independently.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

// ClientRepository stores buyers.
type ClientRepository struct {
	mu      sync.RWMutex
	byID    map[string]domain.Client
	byPhone map[string]domain.Client
}

// NewClientRepository seeds the registered clients.
func NewClientRepository(seed []domain.Client) *ClientRepository {
	r := &ClientRepository{byID: map[string]domain.Client{}, byPhone: map[string]domain.Client{}}
	for _, c := range seed {
		_ = r.Save(context.Background(), c)
	}
	return r
}

func (r *ClientRepository) FindByID(_ context.Context, id string) (domain.Client, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byID[id]
	return c, ok, nil
}

func (r *ClientRepository) FindByPhone(_ context.Context, phone string) (domain.Client, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byPhone[phone]
	return c, ok, nil
}

func (r *ClientRepository) Save(_ context.Context, c domain.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[c.ID] = c
	if c.Phone != "" {
		r.byPhone[c.Phone] = c
	}
	return nil
}

// PrescriptionRepository stores OCR results.
type PrescriptionRepository struct {
	mu   sync.RWMutex
	byID map[string]domain.Prescription
}

func NewPrescriptionRepository() *PrescriptionRepository {
	return &PrescriptionRepository{byID: map[string]domain.Prescription{}}
}

func (r *PrescriptionRepository) Save(_ context.Context, p domain.Prescription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.Items = append([]domain.PrescribedItem(nil), p.Items...)
	r.byID[p.ID] = p
	return nil
}

func (r *PrescriptionRepository) FindByID(_ context.Context, id string) (domain.Prescription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	if !ok {
		return domain.Prescription{}, domain.NotFound("Receta", id)
	}
	p.Items = append([]domain.PrescribedItem(nil), p.Items...)
	return p, nil
}

// OrderRepository stores orders and issues the demo codes.
type OrderRepository struct {
	mu   sync.RWMutex
	byID map[string]domain.Order
	seq  int
}

func NewOrderRepository() *OrderRepository { return &OrderRepository{byID: map[string]domain.Order{}} }

func (r *OrderRepository) Save(_ context.Context, o domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[o.ID] = o.Clone()
	return nil
}

func (r *OrderRepository) FindByID(_ context.Context, id string) (domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.byID[id]
	if !ok {
		return domain.Order{}, domain.NotFound("Pedido", id)
	}
	return o.Clone(), nil
}

func (r *OrderRepository) List(_ context.Context) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Order, 0, len(r.byID))
	for _, o := range r.byID {
		out = append(out, o.Clone())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (r *OrderRepository) NextCode(_ context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	return fmt.Sprintf("DEMO-%03d", r.seq), nil
}

// PaymentRepository stores payment attempts.
type PaymentRepository struct {
	mu   sync.RWMutex
	byID map[string]domain.Payment
}

func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{byID: map[string]domain.Payment{}}
}

func (r *PaymentRepository) Save(_ context.Context, p domain.Payment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[p.ID] = p
	return nil
}

func (r *PaymentRepository) FindByID(_ context.Context, id string) (domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.byID[id]
	if !ok {
		return domain.Payment{}, domain.NotFound("Pago", id)
	}
	return p, nil
}

func (r *PaymentRepository) ListByOrder(_ context.Context, orderID string) ([]domain.Payment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Payment
	for _, p := range r.byID {
		if p.OrderID == orderID {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

// ConversationRepository stores the chat state per phone. The assistant
// serialises access per message, so values are stored as given.
type ConversationRepository struct {
	mu      sync.RWMutex
	byPhone map[string]domain.Conversation
}

func NewConversationRepository() *ConversationRepository {
	return &ConversationRepository{byPhone: map[string]domain.Conversation{}}
}

func (r *ConversationRepository) Find(_ context.Context, phone string) (domain.Conversation, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byPhone[phone]
	return c, ok, nil
}

func (r *ConversationRepository) Save(_ context.Context, c domain.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byPhone[c.Phone] = c
	return nil
}

func (r *ConversationRepository) Delete(_ context.Context, phone string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byPhone, phone)
	return nil
}

// MessageLog stores the WhatsApp transcript per phone.
type MessageLog struct {
	mu      sync.RWMutex
	byPhone map[string][]domain.Message
}

func NewMessageLog() *MessageLog { return &MessageLog{byPhone: map[string][]domain.Message{}} }

func (l *MessageLog) Append(_ context.Context, phone string, m domain.Message) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.byPhone[phone] = append(l.byPhone[phone], m)
	return nil
}

func (l *MessageLog) List(_ context.Context, phone string) ([]domain.Message, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return append([]domain.Message(nil), l.byPhone[phone]...), nil
}

func (l *MessageLog) Clear(_ context.Context, phone string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.byPhone, phone)
	return nil
}

// NotificationRepository is the outbox.
type NotificationRepository struct {
	mu   sync.RWMutex
	list []domain.Notification
}

func NewNotificationRepository() *NotificationRepository { return &NotificationRepository{} }

func (r *NotificationRepository) Save(_ context.Context, n domain.Notification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.list = append(r.list, n)
	return nil
}

func (r *NotificationRepository) List(_ context.Context, f domain.NotificationFilter) ([]domain.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []domain.Notification
	for _, n := range r.list {
		if f.Channel != "" && n.Channel != f.Channel {
			continue
		}
		if f.OrderID != "" && n.OrderID != f.OrderID {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}
