package domain_test

import (
	"testing"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

func sampleOrder() domain.Order {
	return domain.Order{
		ID: "ord_1", Code: "DEMO-001", Mode: domain.ModePickup, Status: domain.OrderPending,
		ReservationExpiresAt: time.Date(2026, 10, 8, 15, 10, 0, 0, time.UTC),
		Items: []domain.OrderItem{
			{ID: "itm_1", SKU: "PAR-ALFA-500", Brand: "Marca Alfa", PharmacyID: "med-norte", PharmacyName: "Medicity Demo Norte", Quantity: 20, PrescribedQuantity: 20, UnitPrice: 40},
			{ID: "itm_2", SKU: "AMX-BETA-500", Brand: "Marca Beta", PharmacyID: "eco-norte", PharmacyName: "Económicas Demo Norte", Quantity: 21, PrescribedQuantity: 21, UnitPrice: 55, RequiresPrescription: true},
		},
	}
}

func wantCode(t *testing.T, err error, code domain.ErrorCode) {
	t.Helper()
	if !domain.IsCode(err, code) {
		t.Fatalf("want error %s, got %v", code, err)
	}
}

func TestQuantityChangeRules(t *testing.T) {
	o := sampleOrder()
	if _, delta, err := o.QuantityChange("itm_1", 25); err != nil || delta != 5 {
		t.Fatalf("OTC increase: delta=%d err=%v", delta, err)
	}
	_, _, err := o.QuantityChange("itm_2", 22)
	wantCode(t, err, domain.CodeRxIncreaseNotAllowed)
	if _, delta, err := o.QuantityChange("itm_2", 20); err != nil || delta != -1 {
		t.Fatalf("Rx decrease: delta=%d err=%v", delta, err)
	}
	if o.Items[1].CanIncrease() {
		t.Fatal("Rx item at the prescribed quantity must not increase")
	}
	_, _, err = o.QuantityChange("itm_1", -1)
	wantCode(t, err, domain.CodeValidation)
	_, _, err = o.QuantityChange("nope", 1)
	wantCode(t, err, domain.CodeNotFound)
}

func TestRxCanGoBackUpToThePrescribedQuantity(t *testing.T) {
	o := sampleOrder()
	o.ApplyQuantity("itm_2", 15)
	rx, _ := o.Item("itm_2")
	if !rx.CanIncrease() {
		t.Fatal("Rx item below the prescribed quantity must be able to increase")
	}
	if _, delta, err := o.QuantityChange("itm_2", 21); err != nil || delta != 6 {
		t.Fatalf("Rx back to the prescribed quantity: delta=%d err=%v", delta, err)
	}
	_, _, err := o.QuantityChange("itm_2", 22)
	wantCode(t, err, domain.CodeRxIncreaseNotAllowed)
	o.ApplyQuantity("itm_2", 21)
	if rx, _ := o.Item("itm_2"); rx.CanIncrease() {
		t.Fatal("Rx item back at the prescribed quantity must stop increasing")
	}
}

func TestApplyQuantityZeroKeepsLineUntilPaid(t *testing.T) {
	o := sampleOrder()
	o.ApplyQuantity("itm_2", 0)
	rx, err := o.Item("itm_2")
	if err != nil || rx.Quantity != 0 || len(o.Items) != 2 {
		t.Fatalf("want itm_2 kept at zero, got %+v err=%v", o.Items, err)
	}
	if rx.CanDecrease() || !rx.CanIncrease() {
		t.Fatal("a line at zero can only increase")
	}
	if _, delta, err := o.QuantityChange("itm_2", 21); err != nil || delta != 21 {
		t.Fatalf("Rx back from zero to the prescribed quantity: delta=%d err=%v", delta, err)
	}
	_, _, err = o.QuantityChange("itm_2", 22)
	wantCode(t, err, domain.CodeRxIncreaseNotAllowed)
	if o.IsEmpty() {
		t.Fatal("cart with itm_1 left is not empty")
	}
	o.MarkPaid(time.Now())
	if len(o.Items) != 1 || o.Items[0].ID != "itm_1" {
		t.Fatalf("paying must drop the lines at zero, got %+v", o.Items)
	}
	empty := sampleOrder()
	empty.ApplyQuantity("itm_1", 0)
	empty.ApplyQuantity("itm_2", 0)
	if !empty.IsEmpty() {
		t.Fatal("cart with every line at zero is empty")
	}
}

func TestTotals(t *testing.T) {
	o := sampleOrder()
	if o.Subtotal() != 20*40+21*55 {
		t.Fatalf("subtotal %d", o.Subtotal())
	}
	if o.Total() != o.Subtotal() {
		t.Fatalf("pickup total must not add a fee")
	}
	o.Mode, o.DeliveryFee = domain.ModeDelivery, domain.DefaultDeliveryFee
	if o.Total() != o.Subtotal()+250 {
		t.Fatalf("delivery total %d", o.Total())
	}
	if o.Total().Format() != "USD 22,05" {
		t.Fatalf("format %s", o.Total().Format())
	}
	o.Items = nil
	if o.Total() != 0 {
		t.Fatalf("empty delivery cart must cost 0")
	}
}

func TestEnsurePayable(t *testing.T) {
	now := time.Date(2026, 10, 8, 15, 0, 0, 0, time.UTC)
	o := sampleOrder()
	if err := o.EnsurePayable(now); err != nil {
		t.Fatal(err)
	}
	wantCode(t, o.EnsurePayable(now.Add(time.Hour)), domain.CodeReservationExpired)
	empty := sampleOrder()
	empty.Items = nil
	wantCode(t, empty.EnsurePayable(now), domain.CodeEmptyCart)
	paid := sampleOrder()
	paid.MarkPaid(now)
	wantCode(t, paid.EnsurePayable(now), domain.CodeInvalidState)
}

func TestMarkPaidIsIdempotent(t *testing.T) {
	now := time.Date(2026, 10, 8, 15, 0, 0, 0, time.UTC)
	o := sampleOrder()
	o.MarkPaid(now)
	o.MarkPaid(now.Add(time.Hour))
	if o.Status != domain.OrderPaid || !o.PaidAt.Equal(now) {
		t.Fatalf("status %s paidAt %v", o.Status, o.PaidAt)
	}
	if len(o.Fulfillments) != 2 || o.Fulfillments[0].ETAMinutes != 20 || !o.Fulfillments[0].ETA.Equal(now.Add(20*time.Minute)) {
		t.Fatalf("fulfillments %+v", o.Fulfillments)
	}
}

func TestPickupStatusFollowsPharmacies(t *testing.T) {
	now := time.Date(2026, 10, 8, 15, 0, 0, 0, time.UTC)
	o := sampleOrder()
	o.MarkPaid(now)
	step := func(ph string, s domain.FulfillmentStatus, want domain.OrderStatus) {
		t.Helper()
		if err := o.AdvancePharmacy(ph, s, now); err != nil {
			t.Fatal(err)
		}
		if o.Status != want {
			t.Fatalf("after %s→%s want %s got %s", ph, s, want, o.Status)
		}
	}
	step("med-norte", domain.FulfillmentPreparing, domain.OrderPreparing)
	step("med-norte", domain.FulfillmentReady, domain.OrderPreparing)
	step("eco-norte", domain.FulfillmentReady, domain.OrderReady)
	wantCode(t, o.AdvancePharmacy("eco-norte", domain.FulfillmentPreparing, now), domain.CodeInvalidState)
	step("med-norte", domain.FulfillmentPickedUp, domain.OrderReady)
	step("eco-norte", domain.FulfillmentPickedUp, domain.OrderDelivered)
}

func TestDeliveryTransitions(t *testing.T) {
	now := time.Date(2026, 10, 8, 15, 0, 0, 0, time.UTC)
	o := sampleOrder()
	o.Mode = domain.ModeDelivery
	o.Delivery = &domain.Delivery{Status: domain.DeliveryPending}
	wantCode(t, o.AdvanceDelivery(domain.DeliveryDispatched, now), domain.CodeInvalidState)
	o.MarkPaid(now)
	if o.Delivery.Status != domain.DeliveryConsolidating || o.Delivery.ETAMinutes != 45 {
		t.Fatalf("delivery after pay %+v", o.Delivery)
	}
	if err := o.AdvanceDelivery(domain.DeliveryDispatched, now); err != nil || o.Status != domain.OrderDispatched || o.Delivery.ETAMinutes != 15 {
		t.Fatalf("dispatch: %v %+v", err, o.Delivery)
	}
	if err := o.AdvanceDelivery(domain.DeliveryDelivered, now); err != nil || o.Status != domain.OrderDelivered {
		t.Fatalf("deliver: %v %s", err, o.Status)
	}
}
