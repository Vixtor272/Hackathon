// Package fulfillment is what cashiers and couriers use to move a paid order
// forward; every step tells the client the new ETA by WhatsApp.
package fulfillment

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Service implements ports.FulfillmentService.
type Service struct {
	orders   ports.OrderRepository
	notifier ports.Notifier
	clock    ports.Clock
}

// New wires the fulfilment use cases.
func New(orders ports.OrderRepository, notifier ports.Notifier, clock ports.Clock) *Service {
	return &Service{orders: orders, notifier: notifier, clock: clock}
}

// List returns paid orders for one pharmacy, for the courier, or all.
func (s *Service) List(ctx context.Context, f ports.FulfillmentFilter) ([]domain.Order, error) {
	all, err := s.orders.List(ctx)
	if err != nil {
		return nil, err
	}
	var out []domain.Order
	for _, o := range all {
		if !o.IsPaid() {
			continue
		}
		if f.PharmacyID != "" && o.FulfillmentFor(f.PharmacyID) == nil {
			continue
		}
		if f.Role == "courier" && o.Mode != domain.ModeDelivery {
			continue
		}
		out = append(out, o)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// AdvancePharmacy is the cashier pressing "En preparación" / "Listo para recoger".
func (s *Service) AdvancePharmacy(ctx context.Context, orderID, pharmacyID string, status domain.FulfillmentStatus) (domain.Order, error) {
	o, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return domain.Order{}, err
	}
	now := s.clock.Now()
	if err := o.AdvancePharmacy(pharmacyID, status, now); err != nil {
		return domain.Order{}, err
	}
	if err := s.orders.Save(ctx, o); err != nil {
		return domain.Order{}, err
	}
	f := o.FulfillmentFor(pharmacyID)
	_ = s.notifier.Client(ctx, o.Phone, o.ID, pharmacyMessage(o, *f, now))
	return o, nil
}

// AdvanceDelivery is the courier pressing "En reparto" / "Entregado".
func (s *Service) AdvanceDelivery(ctx context.Context, orderID string, status domain.DeliveryStatus) (domain.Order, error) {
	o, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return domain.Order{}, err
	}
	now := s.clock.Now()
	if err := o.AdvanceDelivery(status, now); err != nil {
		return domain.Order{}, err
	}
	if err := s.orders.Save(ctx, o); err != nil {
		return domain.Order{}, err
	}
	switch status {
	case domain.DeliveryDispatched:
		_ = s.notifier.Courier(ctx, o.Delivery.Courier.Name, o.ID, fmt.Sprintf("Pedido %s en reparto", o.Code), "Dirección: "+o.DeliveryAddress)
		_ = s.notifier.Client(ctx, o.Phone, o.ID, fmt.Sprintf("🛵 Tu pedido %s está en reparto con %s. Llegará en aproximadamente %d minutos a %s.",
			o.Code, o.Delivery.Courier.Name, MinutesUntil(o.Delivery.ETA, now), o.DeliveryAddress))
	case domain.DeliveryDelivered:
		_ = s.notifier.Client(ctx, o.Phone, o.ID, fmt.Sprintf("🏠 Tu pedido %s fue entregado. ¡Gracias por comprar con Farmi!", o.Code))
	}
	return o, nil
}

func pharmacyMessage(o domain.Order, f domain.Fulfillment, now time.Time) string {
	switch f.Status {
	case domain.FulfillmentPreparing:
		if o.Mode == domain.ModeDelivery {
			return fmt.Sprintf("👩‍⚕️ Estamos preparando los productos de tu pedido %s para el envío a domicilio.", o.Code)
		}
		return fmt.Sprintf("👩‍⚕️ Tu pedido %s está en preparación en %s. Estará listo en aproximadamente %d minutos.",
			o.Code, f.PharmacyName, MinutesUntil(f.ETA, now))
	case domain.FulfillmentReady:
		if o.Mode == domain.ModeDelivery {
			return fmt.Sprintf("📦 Los productos de tu pedido %s ya están listos y se consolidan para el envío.", o.Code)
		}
		return fmt.Sprintf("📦 ¡Tu pedido %s está listo para recoger en %s (%s)! Muestra tu código al retirar.", o.Code, f.PharmacyName, f.Address)
	case domain.FulfillmentPickedUp:
		return fmt.Sprintf("🙌 Gracias por tu compra. Tu pedido %s fue entregado en %s.", o.Code, f.PharmacyName)
	}
	return fmt.Sprintf("Tu pedido %s cambió de estado en %s.", o.Code, f.PharmacyName)
}

// MinutesUntil renders an ETA as "minutes from now", never below 1.
func MinutesUntil(eta *time.Time, now time.Time) int {
	if eta == nil {
		return 0
	}
	m := int(math.Ceil(eta.Sub(now).Minutes()))
	if m < 1 {
		return 1
	}
	return m
}
