package order_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/testkit"
)

func TestExpiryReleasesStockAndRenewReserves(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	o := app.PickupOrder(t)
	item := o.Items[0]
	app.Clock.Advance(11 * time.Minute)
	expired, err := app.OrderSvc.ExpireReservations(ctx)
	if err != nil || len(expired) != 1 || expired[0].Status != domain.OrderExpired {
		t.Fatalf("expire: %v %+v", err, expired)
	}
	if _, reserved := app.Stock(t, item.PharmacyID, item.SKU); reserved != 0 {
		t.Fatalf("reserved after expiry %d", reserved)
	}
	if !strings.Contains(app.TranscriptText(t, "+593991111111"), "venció") {
		t.Fatalf("client not told about expiry")
	}
	renewed, err := app.OrderSvc.Renew(ctx, o.ID)
	if err != nil || renewed.Status != domain.OrderPending || !renewed.ReservationActive(app.Clock.Now()) {
		t.Fatalf("renew: %v %+v", err, renewed.Status)
	}
	if _, reserved := app.Stock(t, item.PharmacyID, item.SKU); reserved != item.Quantity {
		t.Fatalf("reserved after renew %d", reserved)
	}
}

func TestCancelReleasesAndNotifies(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	o := app.PickupOrder(t)
	cancelled, err := app.OrderSvc.Cancel(ctx, o.ID)
	if err != nil || cancelled.Status != domain.OrderCancelled {
		t.Fatalf("cancel: %v", err)
	}
	for _, it := range o.Items {
		if _, reserved := app.Stock(t, it.PharmacyID, it.SKU); reserved != 0 {
			t.Fatalf("%s still reserved %d", it.SKU, reserved)
		}
	}
	if !strings.Contains(app.TranscriptText(t, "+593991111111"), "cancelado") {
		t.Fatalf("client not told about cancellation")
	}
	if _, err := app.OrderSvc.Cancel(ctx, o.ID); !domain.IsCode(err, domain.CodeInvalidState) {
		t.Fatalf("second cancel want INVALID_STATE, got %v", err)
	}
}

func TestIncreaseIsBoundedByStock(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	o := app.PickupOrder(t)
	var otc domain.OrderItem
	for _, it := range o.Items {
		if !it.RequiresPrescription {
			otc = it
			break
		}
	}
	if _, err := app.OrderSvc.ChangeQuantity(ctx, o.ID, otc.ID, 10_000); !domain.IsCode(err, domain.CodeInsufficientStock) {
		t.Fatalf("want INSUFFICIENT_STOCK, got %v", err)
	}
	after, err := app.OrderSvc.Get(ctx, o.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := after.Item(otc.ID); got.Quantity != otc.Quantity {
		t.Fatalf("quantity changed despite failure: %d", got.Quantity)
	}
}
