// Package ports declares the interfaces of the hexagon. Driving ports (what the
// application offers) live in driving.go; driven ports (what it needs from the
// outside world: storage, OCR, the company's API, payment rails, WhatsApp)
// live here. Adapters implement them; services depend only on them.
package ports

import (
	"context"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

// PrescriptionExtractor is the OCR engine.
type PrescriptionExtractor interface {
	Extract(ctx context.Context, mediaID string) (domain.Prescription, error)
	ListMedia(ctx context.Context) ([]domain.SampleMedia, error)
}

// DoctorRegistry is the medical registry consulted during validation.
type DoctorRegistry interface {
	FindByRegistryID(ctx context.Context, registryID string) (domain.Doctor, bool, error)
}

// CatalogAPI is the company's external product/pharmacy API.
type CatalogAPI interface {
	ListZones(ctx context.Context) ([]domain.Zone, error)
	FindZone(ctx context.Context, id string) (domain.Zone, error)
	ListPharmacies(ctx context.Context, zoneID string) ([]domain.Pharmacy, error) // "" = every zone
	FindPharmacy(ctx context.Context, id string) (domain.Pharmacy, error)
	ListMedicines(ctx context.Context) ([]domain.Medicine, error)
	ProductsByMedicine(ctx context.Context, medicineKey string) ([]domain.Product, error)
	FindProduct(ctx context.Context, sku string) (domain.Product, error)
}

// InventoryRepository holds stock and reservations per pharmacy and SKU.
// Availability = stock − reservations (demo model).
type InventoryRepository interface {
	Available(ctx context.Context, pharmacyID, sku string) (int, error)
	Reserve(ctx context.Context, pharmacyID, sku string, qty int) error
	Release(ctx context.Context, pharmacyID, sku string, qty int) error
	Commit(ctx context.Context, pharmacyID, sku string, qty int) error
	Levels(ctx context.Context) ([]domain.StockLevel, error)
}

// ConversationAI is the language-understanding side of the AI module: it
// reads client replies and maps OCR text onto the catalog.
type ConversationAI interface {
	Interpret(ctx context.Context, text string) domain.Interpretation
	MatchZone(ctx context.Context, text string, zones []domain.Zone) (domain.Zone, bool)
	MatchMedicine(ctx context.Context, item domain.PrescribedItem, medicines []domain.Medicine) (domain.Medicine, bool)
}

// ClientRepository stores registered buyers.
type ClientRepository interface {
	FindByID(ctx context.Context, id string) (domain.Client, bool, error)
	FindByPhone(ctx context.Context, phone string) (domain.Client, bool, error)
	Save(ctx context.Context, c domain.Client) error
}

// PrescriptionRepository stores OCR results.
type PrescriptionRepository interface {
	Save(ctx context.Context, p domain.Prescription) error
	FindByID(ctx context.Context, id string) (domain.Prescription, error)
}

// OrderRepository stores carts/orders.
type OrderRepository interface {
	Save(ctx context.Context, o domain.Order) error
	FindByID(ctx context.Context, id string) (domain.Order, error)
	List(ctx context.Context) ([]domain.Order, error)
	NextCode(ctx context.Context) (string, error)
}

// PaymentRepository stores payment attempts.
type PaymentRepository interface {
	Save(ctx context.Context, p domain.Payment) error
	FindByID(ctx context.Context, id string) (domain.Payment, error)
	ListByOrder(ctx context.Context, orderID string) ([]domain.Payment, error)
}

// ConversationRepository stores the purchase state per phone.
type ConversationRepository interface {
	Find(ctx context.Context, phone string) (domain.Conversation, bool, error)
	Save(ctx context.Context, c domain.Conversation) error
	Delete(ctx context.Context, phone string) error
}

// MessageLog is the WhatsApp transcript per phone.
type MessageLog interface {
	Append(ctx context.Context, phone string, m domain.Message) error
	List(ctx context.Context, phone string) ([]domain.Message, error)
	Clear(ctx context.Context, phone string) error
}

// NotificationRepository is the outbox of every simulated channel.
type NotificationRepository interface {
	Save(ctx context.Context, n domain.Notification) error
	List(ctx context.Context, f domain.NotificationFilter) ([]domain.Notification, error)
}

// MessageSender delivers a WhatsApp message to a phone.
type MessageSender interface {
	Send(ctx context.Context, phone string, m domain.Message) error
}

// PaymentGateway is the payment rail (card processor / DeUna).
type PaymentGateway interface {
	CreateIntent(ctx context.Context, p domain.Payment) (domain.Payment, error)
	ListBanks(ctx context.Context) ([]domain.Bank, error)
}

// Notifier fans order events out to the client, the pharmacies and the courier.
type Notifier interface {
	Client(ctx context.Context, phone, orderID, text string) error
	Pharmacy(ctx context.Context, pharmacyID, orderID, title, body string) error
	Courier(ctx context.Context, courierName, orderID, title, body string) error
}

// PaymentInvalidator voids pending payment attempts when the cart changes.
type PaymentInvalidator interface {
	InvalidatePending(ctx context.Context, orderID string) error
}

// Clock abstracts time so use cases are testable.
type Clock interface {
	Now() time.Time
}

// IDGenerator issues identifiers with a readable prefix ("ord_1").
type IDGenerator interface {
	New(prefix string) string
}
