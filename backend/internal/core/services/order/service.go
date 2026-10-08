// Package order owns the cart: stock re-check, 10-minute reservation, web
// page quantity rules, cancellation, expiry and renewal.
package order

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Config tunes the demo.
type Config struct {
	ReservationTTL time.Duration
	WebBaseURL     string
}

// Service implements ports.OrderService.
type Service struct {
	orders    ports.OrderRepository
	inventory ports.InventoryRepository
	payments  ports.PaymentInvalidator
	notifier  ports.Notifier
	clock     ports.Clock
	ids       ports.IDGenerator
	cfg       Config
	mu        sync.Mutex // serialises reserve/release so stock never goes negative
}

// New wires the order use cases.
func New(orders ports.OrderRepository, inventory ports.InventoryRepository, payments ports.PaymentInvalidator,
	notifier ports.Notifier, clock ports.Clock, ids ports.IDGenerator, cfg Config) *Service {
	if cfg.ReservationTTL <= 0 {
		cfg.ReservationTTL = 10 * time.Minute
	}
	return &Service{orders: orders, inventory: inventory, payments: payments, notifier: notifier, clock: clock, ids: ids, cfg: cfg}
}

// Create re-checks stock, reserves every chosen brand and opens a pending order.
func (s *Service) Create(ctx context.Context, in ports.CreateOrderInput) (domain.Order, error) {
	if len(in.Selections) == 0 {
		return domain.Order{}, domain.NewError(domain.CodeEmptyCart, "El carrito está vacío")
	}
	if in.Mode == domain.ModeDelivery && in.DeliveryAddress == "" {
		return domain.Order{}, domain.NewError(domain.CodeValidation, "Falta la dirección de entrega")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	var held []domain.BrandOption
	for _, sel := range in.Selections {
		if err := s.inventory.Reserve(ctx, sel.PharmacyID, sel.Product.SKU, sel.Quantity); err != nil {
			s.releaseSelections(ctx, held)
			return domain.Order{}, reserveError(err, sel)
		}
		held = append(held, sel)
	}
	code, err := s.orders.NextCode(ctx)
	if err != nil {
		s.releaseSelections(ctx, held)
		return domain.Order{}, err
	}
	now := s.clock.Now()
	o := domain.Order{
		ID: s.ids.New("ord"), Code: code, ClientID: in.ClientID, ClientName: in.ClientName, Phone: in.Phone,
		PrescriptionID: in.PrescriptionID, Mode: in.Mode, Zone: in.Zone, DeliveryAddress: in.DeliveryAddress,
		Status: domain.OrderPending, ReservationExpiresAt: now.Add(s.cfg.ReservationTTL), CreatedAt: now,
	}
	for _, sel := range in.Selections {
		o.Items = append(o.Items, s.itemFrom(sel))
	}
	if in.Mode == domain.ModeDelivery {
		o.DeliveryFee = domain.DefaultDeliveryFee
		d := domain.Delivery{Status: domain.DeliveryPending}
		if in.Courier != nil {
			d.Courier = *in.Courier
		}
		o.Delivery = &d
	}
	if err := s.orders.Save(ctx, o); err != nil {
		s.releaseSelections(ctx, held)
		return domain.Order{}, err
	}
	return o, nil
}

// Get loads an order.
func (s *Service) Get(ctx context.Context, id string) (domain.Order, error) {
	return s.orders.FindByID(ctx, id)
}

// ChangeQuantity applies the web page rules: OTC may grow if stock allows,
// prescription items may only shrink, zero removes, totals and reservation
// are recalculated and any pending payment is voided.
func (s *Service) ChangeQuantity(ctx context.Context, orderID, itemID string, qty int) (domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return domain.Order{}, err
	}
	if err := o.EnsureEditable(s.clock.Now()); err != nil {
		return domain.Order{}, err
	}
	item, delta, err := o.QuantityChange(itemID, qty)
	if err != nil {
		return domain.Order{}, err
	}
	switch {
	case delta > 0:
		if err := s.inventory.Reserve(ctx, item.PharmacyID, item.SKU, delta); err != nil {
			return domain.Order{}, domain.NewError(domain.CodeInsufficientStock,
				"No hay stock adicional de %s en %s; se conserva la cantidad %d", item.Brand, item.PharmacyName, item.Quantity)
		}
	case delta < 0:
		if err := s.inventory.Release(ctx, item.PharmacyID, item.SKU, -delta); err != nil {
			return domain.Order{}, err
		}
	}
	o.ApplyQuantity(itemID, qty)
	if err := s.voidPayment(ctx, &o); err != nil {
		return domain.Order{}, err
	}
	if err := s.orders.Save(ctx, o); err != nil {
		return domain.Order{}, err
	}
	return o, nil
}

// Cancel releases the reserved units of a pending order.
func (s *Service) Cancel(ctx context.Context, id string) (domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	wasPending := o.Status == domain.OrderPending
	if err := o.Cancel(); err != nil {
		return domain.Order{}, err
	}
	if wasPending {
		s.releaseItems(ctx, o)
	}
	if err := s.voidPayment(ctx, &o); err != nil {
		return domain.Order{}, err
	}
	if err := s.orders.Save(ctx, o); err != nil {
		return domain.Order{}, err
	}
	_ = s.notifier.Client(ctx, o.Phone, o.ID,
		fmt.Sprintf("❌ Tu pedido %s fue cancelado y las unidades reservadas fueron liberadas.", o.Code))
	return o, nil
}

// Renew re-checks availability for an expired order and reserves again.
func (s *Service) Renew(ctx context.Context, id string) (domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	now := s.clock.Now()
	if o.ReservationActive(now) {
		return o, nil
	}
	if o.Status == domain.OrderPending { // lapsed but the worker has not run yet
		s.releaseItems(ctx, o)
		o.Expire()
	}
	if o.Status != domain.OrderExpired {
		return domain.Order{}, domain.NewError(domain.CodeInvalidState, "El pedido %s no puede renovarse (estado %s)", o.Code, o.Status)
	}
	var held []domain.OrderItem
	for _, it := range o.Items {
		if err := s.inventory.Reserve(ctx, it.PharmacyID, it.SKU, it.Quantity); err != nil {
			for _, h := range held {
				_ = s.inventory.Release(ctx, h.PharmacyID, h.SKU, h.Quantity)
			}
			_ = s.orders.Save(ctx, o)
			return domain.Order{}, domain.NewError(domain.CodeInsufficientStock,
				"Ya no hay disponibilidad de %s en %s", it.Brand, it.PharmacyName)
		}
		held = append(held, it)
	}
	if err := o.Renew(now, s.cfg.ReservationTTL); err != nil {
		return domain.Order{}, err
	}
	if err := s.orders.Save(ctx, o); err != nil {
		return domain.Order{}, err
	}
	return o, nil
}

// PaymentOptions returns the rails and the checkout link sent by WhatsApp.
func (s *Service) PaymentOptions(ctx context.Context, id string) (ports.PaymentOptions, error) {
	o, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return ports.PaymentOptions{}, err
	}
	return ports.PaymentOptions{CheckoutURL: s.cfg.WebBaseURL + "/checkout/" + o.ID, Options: domain.PaymentOptions()}, nil
}

// ExpireReservations is run by the background worker: lapsed pending orders
// release their units, void their payment and the client is told how to retry.
func (s *Service) ExpireReservations(ctx context.Context) ([]domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	all, err := s.orders.List(ctx)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	var expired []domain.Order
	for _, o := range all {
		if o.Status != domain.OrderPending || now.Before(o.ReservationExpiresAt) {
			continue
		}
		s.releaseItems(ctx, o)
		o.Expire()
		if err := s.voidPayment(ctx, &o); err != nil {
			return expired, err
		}
		if err := s.orders.Save(ctx, o); err != nil {
			return expired, err
		}
		_ = s.notifier.Client(ctx, o.Phone, o.ID, fmt.Sprintf(
			"⏰ La reserva de tu pedido %s venció y las unidades fueron liberadas. Escribe *continuar* para verificar disponibilidad y generar una nueva reserva.", o.Code))
		expired = append(expired, o)
	}
	return expired, nil
}

func (s *Service) itemFrom(sel domain.BrandOption) domain.OrderItem {
	return domain.OrderItem{
		ID: s.ids.New("itm"), SKU: sel.Product.SKU, Medicine: sel.Medicine, Brand: sel.Product.Brand,
		Presentation: sel.Product.Presentation, UnitLabel: sel.Product.UnitLabel,
		PharmacyID: sel.PharmacyID, PharmacyName: sel.PharmacyName, PharmacyAddress: sel.PharmacyAddress,
		Quantity: sel.Quantity, PrescribedQuantity: sel.Quantity, UnitPrice: sel.Product.UnitPrice,
		RequiresPrescription: sel.Product.RequiresPrescription,
	}
}

func (s *Service) voidPayment(ctx context.Context, o *domain.Order) error {
	if err := s.payments.InvalidatePending(ctx, o.ID); err != nil {
		return err
	}
	o.Payment = nil
	return nil
}

func (s *Service) releaseItems(ctx context.Context, o domain.Order) {
	for _, it := range o.Items {
		_ = s.inventory.Release(ctx, it.PharmacyID, it.SKU, it.Quantity)
	}
}

func (s *Service) releaseSelections(ctx context.Context, held []domain.BrandOption) {
	for _, h := range held {
		_ = s.inventory.Release(ctx, h.PharmacyID, h.Product.SKU, h.Quantity)
	}
}

func reserveError(err error, sel domain.BrandOption) error {
	if domain.IsCode(err, domain.CodeInsufficientStock) {
		return domain.NewError(domain.CodeInsufficientStock, "no hay stock suficiente de %s (%s) en %s",
			sel.Product.Brand, sel.Medicine, sel.PharmacyName)
	}
	return err
}
