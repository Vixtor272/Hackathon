// Package payment runs the simulated rails (card, DeUna). Approval is the
// single place where stock is deducted, and it is idempotent.
package payment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Service implements ports.PaymentService and ports.PaymentInvalidator.
type Service struct {
	payments  ports.PaymentRepository
	orders    ports.OrderRepository
	inventory ports.InventoryRepository
	gateway   ports.PaymentGateway
	notifier  ports.Notifier
	clock     ports.Clock
	ids       ports.IDGenerator
	mu        sync.Mutex // one confirmation at a time keeps approval idempotent
}

// New wires the payment use cases.
func New(payments ports.PaymentRepository, orders ports.OrderRepository, inventory ports.InventoryRepository,
	gateway ports.PaymentGateway, notifier ports.Notifier, clock ports.Clock, ids ports.IDGenerator) *Service {
	return &Service{payments: payments, orders: orders, inventory: inventory, gateway: gateway, notifier: notifier, clock: clock, ids: ids}
}

// Create opens a payment attempt for the order's current total. Any earlier
// pending attempt is voided so only one link is valid at a time.
func (s *Service) Create(ctx context.Context, orderID string, method domain.PaymentMethod) (domain.Payment, error) {
	if !domain.ValidPaymentMethod(method) {
		return domain.Payment{}, domain.NewError(domain.CodeValidation, "Método de pago inválido: %s", method)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	o, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return domain.Payment{}, err
	}
	now := s.clock.Now()
	if err := o.EnsurePayable(now); err != nil {
		return domain.Payment{}, err
	}
	if err := s.invalidatePendingLocked(ctx, orderID); err != nil {
		return domain.Payment{}, err
	}
	p := domain.Payment{ID: s.ids.New("pay"), OrderID: orderID, Method: method, Amount: o.Total(),
		Status: domain.PaymentPending, CreatedAt: now}
	p, err = s.gateway.CreateIntent(ctx, p)
	if err != nil {
		return domain.Payment{}, err
	}
	if err := s.payments.Save(ctx, p); err != nil {
		return domain.Payment{}, err
	}
	ref := p.Ref()
	o.Payment = &ref
	if err := s.orders.Save(ctx, o); err != nil {
		return domain.Payment{}, err
	}
	return p, nil
}

// Get loads a payment with its order.
func (s *Service) Get(ctx context.Context, id string) (domain.Payment, domain.Order, error) {
	p, err := s.payments.FindByID(ctx, id)
	if err != nil {
		return domain.Payment{}, domain.Order{}, err
	}
	o, err := s.orders.FindByID(ctx, p.OrderID)
	if err != nil {
		return domain.Payment{}, domain.Order{}, err
	}
	return p, o, nil
}

// Banks lists the fictitious banks of the DeUna simulator.
func (s *Service) Banks(ctx context.Context) ([]domain.Bank, error) { return s.gateway.ListBanks(ctx) }

// InvalidatePending voids open attempts (cart changed, reservation expired).
func (s *Service) InvalidatePending(ctx context.Context, orderID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.invalidatePendingLocked(ctx, orderID)
}

func (s *Service) invalidatePendingLocked(ctx context.Context, orderID string) error {
	list, err := s.payments.ListByOrder(ctx, orderID)
	if err != nil {
		return err
	}
	for _, p := range list {
		if p.Status != domain.PaymentPending {
			continue
		}
		p.Status = domain.PaymentInvalidated
		if err := s.payments.Save(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

// Confirm records the simulator's outcome. On approval the order becomes
// PAID, stock is deducted exactly once, reservations are consumed and every
// party is notified. Repeating an approved confirmation changes nothing.
func (s *Service) Confirm(ctx context.Context, id string, outcome domain.PaymentOutcome, bankID string) (domain.Payment, domain.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, err := s.payments.FindByID(ctx, id)
	if err != nil {
		return domain.Payment{}, domain.Order{}, err
	}
	o, err := s.orders.FindByID(ctx, p.OrderID)
	if err != nil {
		return domain.Payment{}, domain.Order{}, err
	}
	switch p.Status {
	case domain.PaymentApproved:
		return p, o, nil
	case domain.PaymentInvalidated:
		return domain.Payment{}, domain.Order{}, domain.NewError(domain.CodePaymentInvalidated,
			"El intento de pago ya no es válido: el carrito cambió o la reserva venció. Genera un nuevo intento desde la página de pago")
	case domain.PaymentRejected:
		return domain.Payment{}, domain.Order{}, domain.NewError(domain.CodeInvalidState,
			"Este intento ya fue rechazado; genera uno nuevo para reintentar")
	}
	now := s.clock.Now()
	p.BankID = bankID
	switch outcome {
	case domain.OutcomeRejected:
		p.Status = domain.PaymentRejected
		p.ConfirmedAt = &now
		if err := s.persist(ctx, p, &o); err != nil {
			return domain.Payment{}, domain.Order{}, err
		}
		_ = s.notifier.Client(ctx, o.Phone, o.ID, fmt.Sprintf(
			"❌ El pago del pedido %s fue rechazado. Puedes reintentar desde el enlace mientras la reserva siga vigente.", o.Code))
		return p, o, nil
	case domain.OutcomeApproved:
		if err := o.EnsurePayable(now); err != nil {
			p.Status = domain.PaymentInvalidated
			_ = s.payments.Save(ctx, p)
			return domain.Payment{}, domain.Order{}, domain.NewError(domain.CodePaymentInvalidated, "No se puede aprobar el pago: %s", messageOf(err))
		}
		if p.Amount != o.Total() {
			p.Status = domain.PaymentInvalidated
			_ = s.payments.Save(ctx, p)
			return domain.Payment{}, domain.Order{}, domain.NewError(domain.CodePaymentInvalidated,
				"El total del pedido cambió (%s → %s); genera un nuevo intento", p.Amount.Format(), o.Total().Format())
		}
		o.MarkPaid(now)
		for _, it := range o.Items {
			if err := s.inventory.Commit(ctx, it.PharmacyID, it.SKU, it.Quantity); err != nil {
				return domain.Payment{}, domain.Order{}, err
			}
		}
		p.Status = domain.PaymentApproved
		p.ConfirmedAt = &now
		if err := s.persist(ctx, p, &o); err != nil {
			return domain.Payment{}, domain.Order{}, err
		}
		s.notifyPreparation(ctx, o)
		s.notifyClient(ctx, o)
		return p, o, nil
	}
	return domain.Payment{}, domain.Order{}, domain.NewError(domain.CodeValidation, "Resultado de pago inválido: %s", outcome)
}

func (s *Service) persist(ctx context.Context, p domain.Payment, o *domain.Order) error {
	if err := s.payments.Save(ctx, p); err != nil {
		return err
	}
	ref := p.Ref()
	o.Payment = &ref
	return s.orders.Save(ctx, *o)
}

func (s *Service) notifyPreparation(ctx context.Context, o domain.Order) {
	for _, f := range o.Fulfillments {
		body := itemLines(o.ItemsFor(f.PharmacyID)) + "\nCliente: " + o.ClientName
		title := fmt.Sprintf("Nuevo pedido %s — retiro en farmacia", o.Code)
		if o.Mode == domain.ModeDelivery {
			title = fmt.Sprintf("Preparar pedido %s — consolidación para envío a domicilio", o.Code)
			body += "\nModalidad: domicilio (la empresa consolida el pedido)"
		} else {
			body += "\nModalidad: retiro en farmacia"
		}
		_ = s.notifier.Pharmacy(ctx, f.PharmacyID, o.ID, title, body)
	}
	if o.Mode == domain.ModeDelivery && o.Delivery != nil {
		_ = s.notifier.Courier(ctx, o.Delivery.Courier.Name, o.ID, fmt.Sprintf("Entrega asignada — pedido %s", o.Code),
			fmt.Sprintf("Dirección: %s\nCliente: %s\n%s", o.DeliveryAddress, o.ClientName, itemLines(o.Items)))
	}
}

func (s *Service) notifyClient(ctx context.Context, o domain.Order) {
	var b strings.Builder
	fmt.Fprintf(&b, "✅ Pago aprobado. Tu pedido %s está confirmado.\n", o.Code)
	if o.Mode == domain.ModeDelivery && o.Delivery != nil {
		fmt.Fprintf(&b, "🛵 %s lo llevará a %s. Llegará en aproximadamente %d minutos.",
			o.Delivery.Courier.Name, o.DeliveryAddress, o.Delivery.ETAMinutes)
	} else {
		for _, f := range o.Fulfillments {
			fmt.Fprintf(&b, "• %s (%s): listo para recoger en aproximadamente %d minutos.\n", f.PharmacyName, f.Address, f.ETAMinutes)
		}
		fmt.Fprintf(&b, "Muestra el código %s al retirar.", o.Code)
	}
	_ = s.notifier.Client(ctx, o.Phone, o.ID, b.String())
}

func itemLines(items []domain.OrderItem) string {
	var b strings.Builder
	for _, it := range items {
		fmt.Fprintf(&b, "• %s — %s × %d %s\n", it.Medicine, it.Brand, it.Quantity, it.UnitLabel)
	}
	return strings.TrimRight(b.String(), "\n")
}

func messageOf(err error) string {
	var e *domain.Error
	if errors.As(err, &e) {
		return e.Message
	}
	return err.Error()
}
