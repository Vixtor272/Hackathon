package rest_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/farmaenlace/farmi/internal/adapters/in/rest"
	"github.com/farmaenlace/farmi/internal/testkit"
)

func surfaces(t *testing.T) (customer, company http.Handler) {
	t.Helper()
	app := testkit.New(10 * time.Minute)
	s := rest.Services{Assistant: app.Assistant, Availability: app.Availability, Orders: app.OrderSvc,
		Payments: app.PaymentSvc, Fulfillment: app.Fulfillment, Notifications: app.Notifier, Clock: app.Clock}
	return rest.NewRouter(s, rest.Config{Surface: rest.SurfaceCustomer}),
		rest.NewRouter(s, rest.Config{Surface: rest.SurfaceCompany})
}

// routed reports whether the surface mounts the endpoint (a missing route
// answers 404 "Ruta no encontrada"; a mounted one may still 404 a missing id).
func routed(h http.Handler, method, path string) bool {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(`{"status":"DISPATCHED"}`)))
	return !(rec.Code == http.StatusNotFound && strings.Contains(rec.Body.String(), "Ruta no encontrada"))
}

func TestSurfacesExposeOnlyTheirHalf(t *testing.T) {
	customer, company := surfaces(t)
	cases := []struct {
		method, path      string
		customer, company bool
	}{
		{http.MethodPost, "/api/v1/whatsapp/webhook", true, false},
		{http.MethodGet, "/api/v1/whatsapp/media", true, false},
		{http.MethodPatch, "/api/v1/orders/ord_1/items/itm_1", true, false},
		{http.MethodPost, "/api/v1/payments", true, false},
		{http.MethodGet, "/api/v1/payments/banks", true, false},
		{http.MethodGet, "/api/v1/fulfillment/orders?role=courier", false, true},
		{http.MethodPost, "/api/v1/fulfillment/orders/ord_1/delivery/status", false, true},
		{http.MethodGet, "/api/v1/notifications", false, true},
		{http.MethodGet, "/api/v1/catalog/zones", true, true},
		{http.MethodGet, "/api/v1/orders/ord_1", true, true},
	}
	for _, c := range cases {
		if got := routed(customer, c.method, c.path); got != c.customer {
			t.Errorf("customer %s %s: routed=%v, want %v", c.method, c.path, got, c.customer)
		}
		if got := routed(company, c.method, c.path); got != c.company {
			t.Errorf("company %s %s: routed=%v, want %v", c.method, c.path, got, c.company)
		}
	}
}

func TestHealthReportsTheSurface(t *testing.T) {
	_, company := surfaces(t)
	rec := httptest.NewRecorder()
	company.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))
	var body struct {
		Surface string `json:"surface"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Surface != "company" {
		t.Fatalf("health %d %s", rec.Code, rec.Body.String())
	}
}
