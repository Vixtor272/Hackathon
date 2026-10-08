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

// Config tunes the HTTP layer.
type Config struct {
	StaticDir string            // built Svelte app to serve at "/", empty = API only
	Modules   map[string]string // reported by /health
}

type handlers struct {
	s   Services
	cfg Config
}

// NewRouter mounts the v1 contract.
func NewRouter(s Services, cfg Config) http.Handler {
	h := &handlers{s: s, cfg: cfg}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", h.health)

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

	mux.HandleFunc("GET /api/v1/fulfillment/orders", h.fulfillmentOrders)
	mux.HandleFunc("POST /api/v1/fulfillment/orders/{id}/pharmacies/{pharmacyId}/status", h.pharmacyStatus)
	mux.HandleFunc("POST /api/v1/fulfillment/orders/{id}/delivery/status", h.deliveryStatus)

	mux.HandleFunc("GET /api/v1/notifications", h.notifications)
	mux.HandleFunc("GET /api/v1/clients/{id}", h.getClient)

	mux.HandleFunc("/api/", h.apiNotFound)
	if cfg.StaticDir != "" {
		mux.Handle("/", spaHandler(cfg.StaticDir))
	} else {
		mux.HandleFunc("/", h.apiRoot)
	}
	return logging(cors(mux))
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

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond))
	})
}
