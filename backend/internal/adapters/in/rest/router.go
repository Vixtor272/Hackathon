package rest

import (
	"log"
	"net/http"
	"time"

	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Services are the driving ports the REST adapter exposes.
type Services struct {
	Assistant     ports.Assistant
	OCR           ports.OCRService
	Validator     ports.PrescriptionValidator
	Availability  ports.AvailabilityService
	Orders        ports.OrderService
	Payments      ports.PaymentService
	Fulfillment   ports.FulfillmentService
	Notifications ports.NotificationService
	Clients       ports.ClientService
	Clock         ports.Clock
}

// Surface is one half of the API. The customer experience and the company's
// back office listen on different ports so neither can reach the other's
// endpoints: a client's browser never sees the cashier board, and the
// cashier app never talks to the WhatsApp channel.
type Surface string

const (
	// SurfaceCustomer: WhatsApp channel, checkout cart, payments (card / DeUna).
	SurfaceCustomer Surface = "customer"
	// SurfaceCompany: cashier and courier board, notifications outbox, clients.
	SurfaceCompany Surface = "company"
)

// Config tunes the HTTP layer.
type Config struct {
	Surface   Surface           // which half of the API this listener exposes
	StaticDir string            // built Svelte app to serve at "/", empty = API only
	Modules   map[string]string // reported by /health
}

type handlers struct {
	s   Services
	cfg Config
}

// NewRouter mounts the v1 contract of one surface.
func NewRouter(s Services, cfg Config) http.Handler {
	h := &handlers{s: s, cfg: cfg}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", h.health)
	switch cfg.Surface {
	case SurfaceCustomer:
		h.mountCustomer(mux)
	case SurfaceCompany:
		h.mountCompany(mux)
	default:
		panic("rest: unknown surface " + string(cfg.Surface))
	}

	mux.HandleFunc("/api/", h.apiNotFound)
	if cfg.StaticDir != "" {
		mux.Handle("/", spaHandler(cfg.StaticDir))
	} else {
		mux.HandleFunc("/", h.apiRoot)
	}
	return logging(cfg.Surface, cors(mux))
}

// mountCustomer: everything the client touches, from the chat to the payment.
func (h *handlers) mountCustomer(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/whatsapp/webhook", h.webhook)
	mux.HandleFunc("GET /api/v1/whatsapp/conversations/{phone}/messages", h.transcript)
	mux.HandleFunc("DELETE /api/v1/whatsapp/conversations/{phone}", h.resetConversation)
	mux.HandleFunc("GET /api/v1/whatsapp/media", h.sampleMedia)

	mux.HandleFunc("POST /api/v1/ocr/extract", h.extract)
	mux.HandleFunc("POST /api/v1/prescriptions/validate", h.validate)

	mux.HandleFunc("GET /api/v1/catalog/zones", h.zones)
	mux.HandleFunc("GET /api/v1/catalog/pharmacies", h.pharmacies)
	mux.HandleFunc("POST /api/v1/catalog/availability", h.availability)
	mux.HandleFunc("POST /api/v1/catalog/brands", h.brands)

	mux.HandleFunc("GET /api/v1/orders/{id}", h.getOrder)
	mux.HandleFunc("PATCH /api/v1/orders/{id}/items/{itemId}", h.changeQuantity)
	mux.HandleFunc("POST /api/v1/orders/{id}/cancel", h.cancelOrder)
	mux.HandleFunc("POST /api/v1/orders/{id}/renew", h.renewOrder)
	mux.HandleFunc("GET /api/v1/orders/{id}/payment-options", h.paymentOptions)

	mux.HandleFunc("POST /api/v1/payments", h.createPayment)
	mux.HandleFunc("GET /api/v1/payments/banks", h.banks)
	mux.HandleFunc("GET /api/v1/payments/{id}", h.getPayment)
	mux.HandleFunc("POST /api/v1/payments/{id}/confirm", h.confirmPayment)
}

// mountCompany: the back office that prepares, dispatches and audits orders.
func (h *handlers) mountCompany(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/catalog/zones", h.zones)
	mux.HandleFunc("GET /api/v1/catalog/pharmacies", h.pharmacies)

	mux.HandleFunc("GET /api/v1/orders/{id}", h.getOrder)

	mux.HandleFunc("GET /api/v1/fulfillment/orders", h.fulfillmentOrders)
	mux.HandleFunc("POST /api/v1/fulfillment/orders/{id}/pharmacies/{pharmacyId}/status", h.pharmacyStatus)
	mux.HandleFunc("POST /api/v1/fulfillment/orders/{id}/delivery/status", h.deliveryStatus)

	mux.HandleFunc("GET /api/v1/notifications", h.notifications)
	mux.HandleFunc("GET /api/v1/clients/{id}", h.getClient)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func logging(surface Surface, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("[%s] %s %s -> %d (%s)", surface, r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond))
	})
}
