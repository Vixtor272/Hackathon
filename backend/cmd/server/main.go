// Command server is the composition root: it builds every adapter, injects
// them into the core services through their ports and starts the HTTP API.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/farmaenlace/farmi/internal/adapters/in/rest"
	"github.com/farmaenlace/farmi/internal/adapters/out/ai/rules"
	catalogmock "github.com/farmaenlace/farmi/internal/adapters/out/catalogapi/mock"
	"github.com/farmaenlace/farmi/internal/adapters/out/clock"
	doctorsmemory "github.com/farmaenlace/farmi/internal/adapters/out/doctors/memory"
	"github.com/farmaenlace/farmi/internal/adapters/out/id"
	ocrmock "github.com/farmaenlace/farmi/internal/adapters/out/ocr/mock"
	paysim "github.com/farmaenlace/farmi/internal/adapters/out/payment/simulator"
	"github.com/farmaenlace/farmi/internal/adapters/out/storage/memory"
	wasim "github.com/farmaenlace/farmi/internal/adapters/out/whatsapp/simulator"
	"github.com/farmaenlace/farmi/internal/core/services/catalog"
	"github.com/farmaenlace/farmi/internal/core/services/client"
	"github.com/farmaenlace/farmi/internal/core/services/farmi"
	"github.com/farmaenlace/farmi/internal/core/services/fulfillment"
	"github.com/farmaenlace/farmi/internal/core/services/notify"
	"github.com/farmaenlace/farmi/internal/core/services/ocr"
	"github.com/farmaenlace/farmi/internal/core/services/order"
	"github.com/farmaenlace/farmi/internal/core/services/payment"
	"github.com/farmaenlace/farmi/internal/core/services/prescription"
	"github.com/farmaenlace/farmi/internal/demo"
)

type config struct {
	port           string
	webBaseURL     string
	staticDir      string
	reservationTTL time.Duration
	expiryTick     time.Duration
}

func loadConfig() config {
	c := config{
		port:           env("PORT", "8080"),
		staticDir:      env("STATIC_DIR", ""),
		reservationTTL: envDuration("RESERVATION_TTL", 10*time.Minute),
		expiryTick:     envDuration("EXPIRY_TICK", 5*time.Second),
	}
	switch c.staticDir {
	case "none":
		c.staticDir = ""
	case "":
		if _, err := os.Stat(filepath.Join("..", "frontend", "dist", "index.html")); err == nil {
			c.staticDir = filepath.Join("..", "frontend", "dist")
		}
	}
	c.webBaseURL = env("WEB_BASE_URL", "")
	if c.webBaseURL == "" {
		if c.staticDir != "" {
			c.webBaseURL = "http://localhost:" + c.port
		} else {
			c.webBaseURL = "http://localhost:5173"
		}
	}
	return c
}

func main() {
	cfg := loadConfig()
	clk := clock.System{}
	ids := id.NewSequential()

	// Storage (one in-memory store per repository port).
	clients := memory.NewClientRepository(demo.Clients())
	prescriptions := memory.NewPrescriptionRepository()
	orders := memory.NewOrderRepository()
	payments := memory.NewPaymentRepository()
	conversations := memory.NewConversationRepository()
	messages := memory.NewMessageLog()
	notifications := memory.NewNotificationRepository()
	inventory := memory.NewInventoryRepository(demo.Stock())

	// External systems, all simulated.
	catalogAPI := catalogmock.NewClient(catalogmock.Data{Zones: demo.Zones(), Pharmacies: demo.Pharmacies(), Medicines: demo.Medicines(), Products: demo.Products()})
	extractor := ocrmock.NewExtractor(demo.Samples())
	doctors := doctorsmemory.NewRegistry(demo.Doctors())
	ai := rules.New()
	gateway := paysim.NewGateway(cfg.webBaseURL, demo.Banks())
	sender := wasim.NewSender(messages)

	// Core services.
	notifier := notify.New(sender, notifications, clk, ids)
	ocrSvc := ocr.New(extractor, prescriptions, ids)
	validator := prescription.NewValidator(doctors)
	availability := catalog.New(catalogAPI, inventory, ai)
	paymentSvc := payment.New(payments, orders, inventory, gateway, notifier, clk, ids)
	orderSvc := order.New(orders, inventory, paymentSvc, notifier, clk, ids, order.Config{ReservationTTL: cfg.reservationTTL, WebBaseURL: cfg.webBaseURL})
	fulfillmentSvc := fulfillment.New(orders, notifier, clk)
	assistant := farmi.New(farmi.Deps{
		Conversations: conversations, Log: messages, Sender: sender, Clients: clients, OCR: ocrSvc, Validator: validator,
		Availability: availability, Orders: orderSvc, AI: ai, Clock: clk, IDs: ids, Courier: demo.Courier(),
	})

	router := rest.NewRouter(rest.Services{
		Assistant: assistant, OCR: ocrSvc, Validator: validator, Availability: availability, Orders: orderSvc,
		Payments: paymentSvc, Fulfillment: fulfillmentSvc, Notifications: notifier, Clients: client.New(clients), Clock: clk,
	}, rest.Config{StaticDir: cfg.staticDir, Modules: map[string]string{
		"ocr": "mock", "catalogApi": "mock", "payments": "simulator", "whatsapp": "simulator", "storage": "memory", "ai": "rules",
	}})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go expiryWorker(ctx, orderSvc, cfg.expiryTick)

	srv := &http.Server{Addr: ":" + cfg.port, Handler: router, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Printf("Farmi backend listening on http://localhost:%s (web: %s, static: %q, reservation TTL: %s)",
			cfg.port, cfg.webBaseURL, cfg.staticDir, cfg.reservationTTL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Println("Farmi backend stopped")
}

// expiryWorker releases lapsed reservations (section 12 of the flow).
func expiryWorker(ctx context.Context, orders *order.Service, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			expired, err := orders.ExpireReservations(ctx)
			if err != nil {
				log.Printf("expiry worker: %v", err)
			}
			for _, o := range expired {
				log.Printf("reservation expired: %s", o.Code)
			}
		}
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		log.Printf("ignoring invalid %s=%q", key, v)
	}
	return fallback
}
