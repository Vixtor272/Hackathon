package farmi_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
	"github.com/farmaenlace/farmi/internal/testkit"
)

const maria = "+593991111111"

func expect(t *testing.T, r ports.ConversationReply, state domain.ConversationState, substr string) {
	t.Helper()
	txt := testkit.Text(r)
	if r.State != state || !strings.Contains(txt, substr) {
		t.Fatalf("want state %s containing %q, got state %s:\n%s", state, substr, r.State, txt)
	}
}

func TestAlwaysAsksForIDBeforeAnything(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	expect(t, app.SendImage(t, maria, "receta-001"), domain.StateAskID, "cédula")
	expect(t, app.Say(t, maria, "hola"), domain.StateAskID, "cédula")
	expect(t, app.Say(t, maria, "abc"), domain.StateAskID, "10 dígitos")
	expect(t, app.Say(t, maria, "1712345678"), domain.StateAskPrescription, "María")
}

func TestPrescriptionRejections(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	app.Say(t, maria, "1712345678")
	expect(t, app.Say(t, maria, "no tengo foto"), domain.StateAskPrescription, "foto")
	expect(t, app.SendImage(t, maria, "receta-004"), domain.StateAskPrescription, "firma")
	expect(t, app.SendImage(t, maria, "receta-003"), domain.StateAskPrescription, "no está activo")
	expect(t, app.SendImage(t, maria, "receta-006"), domain.StateAskPrescription, "no está registrado")
	expect(t, app.SendImage(t, maria, "receta-005"), domain.StateAskPrescription, "ilegible")
	expect(t, app.SendImage(t, maria, "receta-001"), domain.StateAskZone, "Receta validada")
}

func TestPickupHappyPath(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	app.Say(t, maria, "1712345678")
	app.SendImage(t, maria, "receta-001")
	expect(t, app.Say(t, maria, "2"), domain.StateAskMode, "Económicas Demo Centro + Medicity Demo Centro")
	// Only one way to pick it up, so Farmi goes straight to the brands.
	expect(t, app.Say(t, maria, "retiro"), domain.StateAskBrand, "Paracetamol")
	expect(t, app.Say(t, maria, "1"), domain.StateAskBrand, "Amoxicilina")
	expect(t, app.Say(t, maria, "1"), domain.StateAskBrand, "Loratadina")
	expect(t, app.Say(t, maria, "1"), domain.StateConfirmCart, "Total a pagar")
	r := app.Say(t, maria, "sí")
	expect(t, r, domain.StateAwaitPayment, "/checkout/")
	last := r.Replies[len(r.Replies)-1]
	if last.Type != domain.MessageLink || !strings.HasPrefix(last.Link, testkit.WebBaseURL+"/checkout/") {
		t.Fatalf("want checkout link, got %+v", last)
	}
	tr, _ := app.Assistant.Transcript(ctx, maria)
	o, err := app.OrderSvc.Get(ctx, tr.OrderID)
	if err != nil || len(o.Items) != 3 || len(o.PharmacyIDs()) != 2 {
		t.Fatalf("order %v %+v", err, o)
	}
	expect(t, app.Say(t, maria, "estado"), domain.StateAwaitPayment, "pendiente de pago")
	p, _ := app.PaymentSvc.Create(ctx, o.ID, domain.MethodCard)
	if _, _, err := app.PaymentSvc.Confirm(ctx, p.ID, domain.OutcomeApproved, ""); err != nil {
		t.Fatal(err)
	}
	expect(t, app.Say(t, maria, "estado"), domain.StateCompleted, "pagado")
	if _, err := app.Fulfillment.AdvancePharmacy(ctx, o.ID, "eco-centro", domain.FulfillmentReady); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(app.TranscriptText(t, maria), "listo para recoger") {
		t.Fatalf("client not told the order is ready")
	}
	expect(t, app.Say(t, maria, "hola"), domain.StateAskID, "cédula")
}

func TestDeliveryHappyPathWithNewClient(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	phone := "+593993333333"
	expect(t, app.Say(t, phone, "buenas"), domain.StateAskID, "cédula")
	expect(t, app.Say(t, phone, "1101234567"), domain.StateAskName, "llamas")
	expect(t, app.Say(t, phone, "Carla Demo"), domain.StateAskPrescription, "Carla")
	app.SendImage(t, phone, "receta-002")
	expect(t, app.Say(t, phone, "guayaquil"), domain.StateAskMode, "A domicilio sí podemos")
	expect(t, app.Say(t, phone, "retiro"), domain.StateAskMode, "No hay opciones de retiro")
	expect(t, app.Say(t, phone, "domicilio"), domain.StateAskAddress, "dirección")
	expect(t, app.Say(t, phone, "Av. Demo 123 y Calle 4"), domain.StateAskBrand, "Repartidor")
	app.Say(t, phone, "1")
	app.Say(t, phone, "1")
	expect(t, app.Say(t, phone, "1"), domain.StateConfirmCart, "Envío")
	expect(t, app.Say(t, phone, "si"), domain.StateAwaitPayment, "/checkout/")
	tr, _ := app.Assistant.Transcript(ctx, phone)
	o, _ := app.OrderSvc.Get(ctx, tr.OrderID)
	if o.Mode != domain.ModeDelivery || o.Delivery == nil || o.Delivery.Status != domain.DeliveryPending || o.Total() != o.Subtotal()+domain.DefaultDeliveryFee {
		t.Fatalf("delivery order %+v", o)
	}
	p, _ := app.PaymentSvc.Create(ctx, o.ID, domain.MethodCard)
	_, paid, err := app.PaymentSvc.Confirm(ctx, p.ID, domain.OutcomeApproved, "")
	if err != nil || paid.Delivery.Status != domain.DeliveryConsolidating || paid.Delivery.ETAMinutes != 45 {
		t.Fatalf("paid delivery %v %+v", err, paid.Delivery)
	}
	if _, err := app.Fulfillment.AdvanceDelivery(ctx, o.ID, domain.DeliveryDispatched); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Fulfillment.AdvanceDelivery(ctx, o.ID, domain.DeliveryDelivered); err != nil {
		t.Fatal(err)
	}
	txt := app.TranscriptText(t, phone)
	for _, want := range []string{"Pago aprobado", "en reparto", "fue entregado"} {
		if !strings.Contains(txt, want) {
			t.Fatalf("missing %q in transcript:\n%s", want, txt)
		}
	}
}

func TestExpiredReservationCanBeRenewedFromChat(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	ctx := context.Background()
	app.Say(t, maria, "1712345678")
	app.SendImage(t, maria, "receta-001")
	expect(t, app.Say(t, maria, "1"), domain.StateAskMode, "te sugiero esta farmacia")
	expect(t, app.Say(t, maria, "retiro"), domain.StateAskBrand, "Retiro en: Medicity Demo Norte")
	app.Say(t, maria, "1")
	app.Say(t, maria, "1")
	expect(t, app.Say(t, maria, "1"), domain.StateConfirmCart, "Total a pagar")
	app.Say(t, maria, "sí")
	app.Clock.Advance(11 * time.Minute)
	if _, err := app.OrderSvc.ExpireReservations(ctx); err != nil {
		t.Fatal(err)
	}
	expect(t, app.Say(t, maria, "estado"), domain.StateAwaitPayment, "venció")
	expect(t, app.Say(t, maria, "continuar"), domain.StateAwaitPayment, "nueva reserva")
	expect(t, app.Say(t, maria, "cancelar"), domain.StateCompleted, "nueva compra")
}
