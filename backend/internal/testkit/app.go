// Package testkit wires the whole hexagon with in-memory adapters and a fake
// clock so use cases can be exercised end to end in unit tests.
package testkit

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/farmaenlace/farmi/internal/adapters/out/ai/rules"
	catalogmock "github.com/farmaenlace/farmi/internal/adapters/out/catalogapi/mock"
	doctorsmemory "github.com/farmaenlace/farmi/internal/adapters/out/doctors/memory"
	"github.com/farmaenlace/farmi/internal/adapters/out/id"
	ocrmock "github.com/farmaenlace/farmi/internal/adapters/out/ocr/mock"
	paysim "github.com/farmaenlace/farmi/internal/adapters/out/payment/simulator"
	"github.com/farmaenlace/farmi/internal/adapters/out/storage/memory"
	wasim "github.com/farmaenlace/farmi/internal/adapters/out/whatsapp/simulator"
	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
	"github.com/farmaenlace/farmi/internal/core/services/catalog"
	"github.com/farmaenlace/farmi/internal/core/services/farmi"
	"github.com/farmaenlace/farmi/internal/core/services/fulfillment"
	"github.com/farmaenlace/farmi/internal/core/services/notify"
	"github.com/farmaenlace/farmi/internal/core/services/ocr"
	"github.com/farmaenlace/farmi/internal/core/services/order"
	"github.com/farmaenlace/farmi/internal/core/services/payment"
	"github.com/farmaenlace/farmi/internal/core/services/prescription"
	"github.com/farmaenlace/farmi/internal/demo"
)

// Clock is a controllable ports.Clock.
type Clock struct {
	mu  sync.Mutex
	now time.Time
}

// NewClock starts at a fixed instant.
func NewClock(now time.Time) *Clock { return &Clock{now: now} }

// Now returns the fake time.
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance moves the fake time forward.
func (c *Clock) Advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}

// WebBaseURL used for checkout links in tests.
const WebBaseURL = "http://web.test"

// App is the assembled application.
type App struct {
	Clock         *Clock
	Inventory     *memory.InventoryRepository
	Orders        *memory.OrderRepository
	Payments      *memory.PaymentRepository
	Messages      *memory.MessageLog
	Notifications *memory.NotificationRepository
	OCR           *ocr.Service
	Availability  *catalog.Service
	OrderSvc      *order.Service
	PaymentSvc    *payment.Service
	Fulfillment   *fulfillment.Service
	Notifier      *notify.Notifier
	Assistant     *farmi.Assistant
}

// New assembles the app with the demo dataset.
func New(ttl time.Duration) *App {
	clk := NewClock(time.Date(2026, 10, 8, 15, 0, 0, 0, time.UTC))
	ids := id.NewSequential()
	clients := memory.NewClientRepository(demo.Clients())
	prescriptions := memory.NewPrescriptionRepository()
	orders := memory.NewOrderRepository()
	payments := memory.NewPaymentRepository()
	conversations := memory.NewConversationRepository()
	messages := memory.NewMessageLog()
	notifications := memory.NewNotificationRepository()
	inventory := memory.NewInventoryRepository(demo.Stock())

	api := catalogmock.NewClient(catalogmock.Data{Zones: demo.Zones(), Pharmacies: demo.Pharmacies(), Medicines: demo.Medicines(), Products: demo.Products()})
	ai := rules.New()
	sender := wasim.NewSender(messages)
	notifier := notify.New(sender, notifications, clk, ids)
	ocrSvc := ocr.New(ocrmock.NewExtractor(demo.Samples()), prescriptions, ids)
	validator := prescription.NewValidator(doctorsmemory.NewRegistry(demo.Doctors()))
	availability := catalog.New(api, inventory, ai)
	paymentSvc := payment.New(payments, orders, inventory, paysim.NewGateway(WebBaseURL, demo.Banks()), notifier, clk, ids)
	orderSvc := order.New(orders, inventory, paymentSvc, notifier, clk, ids, order.Config{ReservationTTL: ttl, WebBaseURL: WebBaseURL})
	fulfillmentSvc := fulfillment.New(orders, notifier, clk)
	assistant := farmi.New(farmi.Deps{
		Conversations: conversations, Log: messages, Sender: sender, Clients: clients, OCR: ocrSvc, Validator: validator,
		Availability: availability, Orders: orderSvc, AI: ai, Clock: clk, IDs: ids, Courier: demo.Courier(),
	})
	return &App{Clock: clk, Inventory: inventory, Orders: orders, Payments: payments, Messages: messages, Notifications: notifications,
		OCR: ocrSvc, Availability: availability, OrderSvc: orderSvc, PaymentSvc: paymentSvc, Fulfillment: fulfillmentSvc, Notifier: notifier, Assistant: assistant}
}

// Receta runs the OCR on a sample image.
func (a *App) Receta(t testing.TB, mediaID string) domain.Prescription {
	t.Helper()
	rx, err := a.OCR.Extract(context.Background(), mediaID)
	if err != nil {
		t.Fatalf("extract %s: %v", mediaID, err)
	}
	return rx
}

// Say sends a text message to Farmi.
func (a *App) Say(t testing.TB, phone, text string) ports.ConversationReply {
	t.Helper()
	return a.send(t, ports.InboundMessage{From: phone, Type: domain.MessageText, Text: text})
}

// SendImage sends a prescription image to Farmi.
func (a *App) SendImage(t testing.TB, phone, mediaID string) ports.ConversationReply {
	t.Helper()
	return a.send(t, ports.InboundMessage{From: phone, Type: domain.MessageImage, MediaID: mediaID})
}

func (a *App) send(t testing.TB, in ports.InboundMessage) ports.ConversationReply {
	t.Helper()
	r, err := a.Assistant.HandleInbound(context.Background(), in)
	if err != nil {
		t.Fatalf("handle inbound %+v: %v", in, err)
	}
	return r
}

// Stock returns (stock, reserved) for a SKU in a pharmacy.
func (a *App) Stock(t testing.TB, pharmacyID, sku string) (int, int) {
	t.Helper()
	levels, err := a.Inventory.Levels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range levels {
		if l.PharmacyID == pharmacyID && l.SKU == sku {
			return l.Stock, l.Reserved
		}
	}
	return 0, 0
}

// TranscriptText joins every message of a phone.
func (a *App) TranscriptText(t testing.TB, phone string) string {
	t.Helper()
	msgs, err := a.Messages.List(context.Background(), phone)
	if err != nil {
		t.Fatal(err)
	}
	out := ""
	for _, m := range msgs {
		out += m.Text + "\n"
	}
	return out
}

// Text joins the replies of one turn.
func Text(r ports.ConversationReply) string {
	out := ""
	for _, m := range r.Replies {
		out += m.Text + "\n"
	}
	return out
}

// PickupOrder drives the order service directly: receta-001, zona norte,
// first option, first brand of every medicine.
func (a *App) PickupOrder(t testing.TB) domain.Order {
	t.Helper()
	ctx := context.Background()
	rx := a.Receta(t, "receta-001")
	av, err := a.Availability.FindOptions(ctx, "uio-norte", rx)
	if err != nil {
		t.Fatal(err)
	}
	brands, err := a.Availability.BrandOptions(ctx, av.Options[0].Coverage, rx)
	if err != nil {
		t.Fatal(err)
	}
	var selections []domain.BrandOption
	for _, mb := range brands {
		selections = append(selections, mb.Brands[0])
	}
	o, err := a.OrderSvc.Create(ctx, ports.CreateOrderInput{
		Phone: "+593991111111", ClientID: "1712345678", ClientName: "María Pérez", PrescriptionID: rx.ID,
		Mode: domain.ModePickup, Zone: av.Zone, Selections: selections,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	return o
}
