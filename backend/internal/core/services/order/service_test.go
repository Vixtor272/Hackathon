package order_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
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

func TestRxShrinksAndGrowsBackToThePrescribedQuantity(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	o := rxByUnitOrder(t, app)
	var rx domain.OrderItem
	for _, it := range o.Items {
		if it.RequiresPrescription {
			rx = it
			break
		}
	}
	full := rx.Quantity
	if full < 4 {
		t.Fatalf("want a prescription item sold by unit, got %+v", rx)
	}
	if _, err := app.OrderSvc.ChangeQuantity(ctx, o.ID, rx.ID, full-3); err != nil {
		t.Fatalf("reduce: %v", err)
	}
	if _, reserved := app.Stock(t, rx.PharmacyID, rx.SKU); reserved != full-3 {
		t.Fatalf("reserved after reduce %d, want %d", reserved, full-3)
	}
	after, err := app.OrderSvc.ChangeQuantity(ctx, o.ID, rx.ID, full)
	if err != nil {
		t.Fatalf("back to the prescribed quantity: %v", err)
	}
	if got, _ := after.Item(rx.ID); got.Quantity != full || got.CanIncrease() {
		t.Fatalf("after growing back: quantity %d canIncrease %v", got.Quantity, got.CanIncrease())
	}
	if _, reserved := app.Stock(t, rx.PharmacyID, rx.SKU); reserved != full {
		t.Fatalf("reserved after growing back %d, want %d", reserved, full)
	}
	if _, err := app.OrderSvc.ChangeQuantity(ctx, o.ID, rx.ID, full+1); !domain.IsCode(err, domain.CodeRxIncreaseNotAllowed) {
		t.Fatalf("above the prescription want RX_INCREASE_NOT_ALLOWED, got %v", err)
	}
}

// rxByUnitOrder is receta-001 in zona norte with the prescription medicine
// bought by the capsule (not by the box), so its quantity can move by units.
func rxByUnitOrder(t *testing.T, app *testkit.App) domain.Order {
	t.Helper()
	ctx := context.Background()
	rx := app.Receta(t, "receta-001")
	av, err := app.Availability.FindOptions(ctx, "uio-norte", rx)
	if err != nil {
		t.Fatal(err)
	}
	brands, err := app.Availability.BrandOptions(ctx, av.Options[0].Coverage, rx)
	if err != nil {
		t.Fatal(err)
	}
	var selections []domain.BrandOption
	for _, mb := range brands {
		pick := mb.Brands[0]
		for _, b := range mb.Brands {
			if mb.RequiresPrescription && b.Product.SellByUnit {
				pick = b
				break
			}
		}
		selections = append(selections, pick)
	}
	o, err := app.OrderSvc.Create(ctx, ports.CreateOrderInput{
		Phone: "+593991111111", ClientID: "1712345678", ClientName: "María Pérez", PrescriptionID: rx.ID,
		Mode: domain.ModePickup, Zone: av.Zone, Selections: selections,
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	return o
}
