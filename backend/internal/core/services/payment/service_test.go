package payment_test

import (
	"context"
	"testing"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/testkit"
)

func TestApprovalDeductsStockExactlyOnce(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	o := app.PickupOrder(t)
	item := o.Items[0]
	stock, reserved := app.Stock(t, item.PharmacyID, item.SKU)
	if reserved != item.Quantity {
		t.Fatalf("reserved %d want %d", reserved, item.Quantity)
	}
	p, err := app.PaymentSvc.Create(ctx, o.ID, domain.MethodCard)
	if err != nil {
		t.Fatal(err)
	}
	_, paid, err := app.PaymentSvc.Confirm(ctx, p.ID, domain.OutcomeApproved, "")
	if err != nil || paid.Status != domain.OrderPaid {
		t.Fatalf("approve: %v %s", err, paid.Status)
	}
	if s, r := app.Stock(t, item.PharmacyID, item.SKU); s != stock-item.Quantity || r != 0 {
		t.Fatalf("after approval stock=%d reserved=%d (before %d)", s, r, stock)
	}
	_, again, err := app.PaymentSvc.Confirm(ctx, p.ID, domain.OutcomeApproved, "")
	if err != nil || !again.PaidAt.Equal(*paid.PaidAt) {
		t.Fatalf("second confirm: %v", err)
	}
	if s, _ := app.Stock(t, item.PharmacyID, item.SKU); s != stock-item.Quantity {
		t.Fatalf("second confirm deducted stock again: %d", s)
	}
	if notes, _ := app.Notifications.List(ctx, domain.NotificationFilter{Channel: domain.ChannelPharmacy, OrderID: o.ID}); len(notes) != 1 {
		t.Fatalf("want 1 pharmacy notification, got %d", len(notes))
	}
}

func TestCartChangeInvalidatesPendingPayment(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	o := app.PickupOrder(t)
	p, err := app.PaymentSvc.Create(ctx, o.ID, domain.MethodDeUna)
	if err != nil || p.Link != testkit.WebBaseURL+"/deuna/"+p.ID {
		t.Fatalf("create deuna: %v link=%q", err, p.Link)
	}
	var otc domain.OrderItem
	for _, it := range o.Items {
		if !it.RequiresPrescription {
			otc = it
		}
	}
	changed, err := app.OrderSvc.ChangeQuantity(ctx, o.ID, otc.ID, otc.Quantity-1)
	if err != nil || changed.Payment != nil {
		t.Fatalf("change: %v payment=%+v", err, changed.Payment)
	}
	_, _, err = app.PaymentSvc.Confirm(ctx, p.ID, domain.OutcomeApproved, "banco-demo-1")
	if !domain.IsCode(err, domain.CodePaymentInvalidated) {
		t.Fatalf("want PAYMENT_INVALIDATED, got %v", err)
	}
}

func TestRejectedPaymentAllowsRetry(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	o := app.PickupOrder(t)
	p1, _ := app.PaymentSvc.Create(ctx, o.ID, domain.MethodDeUna)
	_, after, err := app.PaymentSvc.Confirm(ctx, p1.ID, domain.OutcomeRejected, "banco-demo-2")
	if err != nil || after.Status != domain.OrderPending || after.Payment == nil || after.Payment.Status != domain.PaymentRejected {
		t.Fatalf("reject: %v %+v", err, after.Payment)
	}
	p2, err := app.PaymentSvc.Create(ctx, o.ID, domain.MethodCard)
	if err != nil {
		t.Fatal(err)
	}
	if _, paid, err := app.PaymentSvc.Confirm(ctx, p2.ID, domain.OutcomeApproved, ""); err != nil || paid.Status != domain.OrderPaid {
		t.Fatalf("retry: %v", err)
	}
	if txt := app.TranscriptText(t, "+593991111111"); !contains(txt, "rechazado") || !contains(txt, "Pago aprobado") {
		t.Fatalf("client messages:\n%s", txt)
	}
}

func TestEmptyOrExpiredCartCannotPay(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	o := app.PickupOrder(t)
	for _, it := range o.Items {
		if _, err := app.OrderSvc.ChangeQuantity(ctx, o.ID, it.ID, 0); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := app.PaymentSvc.Create(ctx, o.ID, domain.MethodCard); !domain.IsCode(err, domain.CodeEmptyCart) {
		t.Fatalf("want EMPTY_CART, got %v", err)
	}
	o2 := app.PickupOrder(t)
	app.Clock.Advance(11 * time.Minute)
	if _, err := app.PaymentSvc.Create(ctx, o2.ID, domain.MethodCard); !domain.IsCode(err, domain.CodeReservationExpired) {
		t.Fatalf("want RESERVATION_EXPIRED, got %v", err)
	}
}

func contains(s, sub string) bool { return len(sub) > 0 && len(s) >= len(sub) && indexOf(s, sub) >= 0 }

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
