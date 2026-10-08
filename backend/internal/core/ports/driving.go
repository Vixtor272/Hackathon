package ports

import (
	"context"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

// InboundMessage is a client message arriving from the WhatsApp channel.
type InboundMessage struct {
	From    string
	Type    domain.MessageType
	Text    string
	MediaID string
}

// ConversationReply is what Farmi answered to one inbound message.
type ConversationReply struct {
	State   domain.ConversationState
	Replies []domain.Message
}

// Transcript is the full chat with one phone.
type Transcript struct {
	Phone    string
	State    domain.ConversationState
	OrderID  string
	Messages []domain.Message
}

// Assistant is Farmi, the WhatsApp sales assistant.
type Assistant interface {
	HandleInbound(ctx context.Context, in InboundMessage) (ConversationReply, error)
	Transcript(ctx context.Context, phone string) (Transcript, error)
	Reset(ctx context.Context, phone string) error
	SampleMedia(ctx context.Context) ([]domain.SampleMedia, error)
}

// OCRService extracts structured prescriptions from images.
type OCRService interface {
	Extract(ctx context.Context, mediaID string) (domain.Prescription, error)
	Find(ctx context.Context, id string) (domain.Prescription, error)
	SampleMedia(ctx context.Context) ([]domain.SampleMedia, error)
}

// PrescriptionValidator applies the completeness and doctor-registry checks.
type PrescriptionValidator interface {
	Validate(ctx context.Context, p domain.Prescription) (domain.ValidationResult, error)
}

// AvailabilityService is the AI module that consults the company's API and
// turns a prescription into purchasable options.
type AvailabilityService interface {
	Zones(ctx context.Context) ([]domain.Zone, error)
	Pharmacies(ctx context.Context, zoneID string) ([]domain.Pharmacy, error)
	FindOptions(ctx context.Context, zoneID string, p domain.Prescription) (domain.Availability, error)
	PlanDelivery(ctx context.Context, p domain.Prescription) (domain.DeliveryPlan, error)
	BrandOptions(ctx context.Context, coverage []domain.Coverage, p domain.Prescription) ([]domain.MedicineBrands, error)
	BrandOptionsAt(ctx context.Context, pharmacyIDs []string, p domain.Prescription) ([]domain.MedicineBrands, error)
}

// CreateOrderInput is everything the chat collected before "confirmar".
type CreateOrderInput struct {
	Phone           string
	ClientID        string
	ClientName      string
	PrescriptionID  string
	Mode            domain.FulfillmentMode
	Zone            domain.Zone
	DeliveryAddress string
	Courier         *domain.Courier
	Selections      []domain.BrandOption
}

// PaymentOptions is the payment menu plus the checkout link.
type PaymentOptions struct {
	CheckoutURL string
	Options     []domain.PaymentOption
}

// OrderService owns the cart, its reservation and its lifecycle.
type OrderService interface {
	Create(ctx context.Context, in CreateOrderInput) (domain.Order, error)
	Get(ctx context.Context, id string) (domain.Order, error)
	ChangeQuantity(ctx context.Context, orderID, itemID string, qty int) (domain.Order, error)
	Cancel(ctx context.Context, id string) (domain.Order, error)
	Renew(ctx context.Context, id string) (domain.Order, error)
	PaymentOptions(ctx context.Context, id string) (PaymentOptions, error)
	ExpireReservations(ctx context.Context) ([]domain.Order, error)
}

// PaymentService runs the simulated payment rails.
type PaymentService interface {
	Create(ctx context.Context, orderID string, method domain.PaymentMethod) (domain.Payment, error)
	Get(ctx context.Context, id string) (domain.Payment, domain.Order, error)
	Banks(ctx context.Context) ([]domain.Bank, error)
	Confirm(ctx context.Context, id string, outcome domain.PaymentOutcome, bankID string) (domain.Payment, domain.Order, error)
}

// FulfillmentFilter narrows the operations board.
type FulfillmentFilter struct {
	PharmacyID string
	Role       string // "courier" lists home deliveries
}

// FulfillmentService is what cashiers and couriers use.
type FulfillmentService interface {
	List(ctx context.Context, f FulfillmentFilter) ([]domain.Order, error)
	AdvancePharmacy(ctx context.Context, orderID, pharmacyID string, status domain.FulfillmentStatus) (domain.Order, error)
	AdvanceDelivery(ctx context.Context, orderID string, status domain.DeliveryStatus) (domain.Order, error)
}

// NotificationService exposes the outbox.
type NotificationService interface {
	List(ctx context.Context, f domain.NotificationFilter) ([]domain.Notification, error)
}

// ClientService exposes registered clients.
type ClientService interface {
	Find(ctx context.Context, id string) (domain.Client, error)
}
