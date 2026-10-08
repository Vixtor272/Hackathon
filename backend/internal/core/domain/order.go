package domain

import (
	"fmt"
	"math"
	"time"
)

// OrderStatus is the lifecycle of a purchase.
type OrderStatus string

const (
	OrderPending    OrderStatus = "PENDING"
	OrderPaid       OrderStatus = "PAID"
	OrderPreparing  OrderStatus = "PREPARING"
	OrderReady      OrderStatus = "READY"
	OrderDispatched OrderStatus = "DISPATCHED"
	OrderDelivered  OrderStatus = "DELIVERED"
	OrderCancelled  OrderStatus = "CANCELLED"
	OrderExpired    OrderStatus = "EXPIRED"
)

// FulfillmentMode is how the client receives the products.
type FulfillmentMode string

const (
	ModePickup   FulfillmentMode = "pickup"
	ModeDelivery FulfillmentMode = "delivery"
)

// FulfillmentStatus is the preparation state inside one pharmacy.
type FulfillmentStatus string

const (
	FulfillmentNotified  FulfillmentStatus = "NOTIFIED"
	FulfillmentPreparing FulfillmentStatus = "PREPARING"
	FulfillmentReady     FulfillmentStatus = "READY"
	FulfillmentPickedUp  FulfillmentStatus = "PICKED_UP"
)

// DeliveryStatus is the state of a home delivery.
type DeliveryStatus string

const (
	DeliveryPending       DeliveryStatus = "PENDING"
	DeliveryConsolidating DeliveryStatus = "CONSOLIDATING"
	DeliveryDispatched    DeliveryStatus = "DISPATCHED"
	DeliveryDelivered     DeliveryStatus = "DELIVERED"
)

// Demo timing and pricing values (all fictitious).
const (
	PickupETA          = 20 * time.Minute
	DeliveryETA        = 45 * time.Minute
	DispatchETA        = 15 * time.Minute
	DefaultDeliveryFee = Money(250)
)

// OrderItem is one chosen brand inside the cart. Quantity is in sale units.
type OrderItem struct {
	ID                   string
	SKU                  string
	Medicine             string
	Brand                string
	Presentation         string
	UnitLabel            string
	PharmacyID           string
	PharmacyName         string
	PharmacyAddress      string
	Quantity             int
	PrescribedQuantity   int
	UnitPrice            Money
	RequiresPrescription bool
}

// Subtotal is price × quantity.
func (i OrderItem) Subtotal() Money { return i.UnitPrice.Times(i.Quantity) }

// CanIncrease: over-the-counter products may grow from the web page (subject
// to stock); prescription products only back up to the prescribed quantity
// after the client reduced them.
func (i OrderItem) CanIncrease() bool {
	return !i.RequiresPrescription || i.Quantity < i.PrescribedQuantity
}

// CanDecrease: any product can be reduced; at zero the line stays in the
// cart, set aside, so the client can add it back.
func (i OrderItem) CanDecrease() bool { return i.Quantity > 0 }

// Fulfillment tracks preparation in one pharmacy.
type Fulfillment struct {
	PharmacyID   string
	PharmacyName string
	Address      string
	Status       FulfillmentStatus
	ETA          *time.Time
	ETAMinutes   int
}

// Courier is the (fictitious) delivery person.
type Courier struct {
	Name  string
	Phone string
}

// Delivery tracks a home delivery.
type Delivery struct {
	Status     DeliveryStatus
	ETA        *time.Time
	ETAMinutes int
	Courier    Courier
}

// PaymentRef is the order's view of its latest payment attempt.
type PaymentRef struct {
	ID     string
	Method PaymentMethod
	Status PaymentStatus
}

// Order is the cart plus its reservation, payment and fulfilment state.
type Order struct {
	ID                   string
	Code                 string
	ClientID             string
	ClientName           string
	Phone                string
	PrescriptionID       string
	Mode                 FulfillmentMode
	Zone                 Zone
	DeliveryAddress      string
	DeliveryFee          Money
	Items                []OrderItem
	Status               OrderStatus
	ReservationExpiresAt time.Time
	Fulfillments         []Fulfillment
	Delivery             *Delivery
	Payment              *PaymentRef
	CreatedAt            time.Time
	PaidAt               *time.Time
}

// Clone returns a deep copy so repositories never share mutable state.
func (o Order) Clone() Order {
	c := o
	c.Items = append([]OrderItem(nil), o.Items...)
	c.Fulfillments = append([]Fulfillment(nil), o.Fulfillments...)
	for i := range c.Fulfillments {
		if o.Fulfillments[i].ETA != nil {
			eta := *o.Fulfillments[i].ETA
			c.Fulfillments[i].ETA = &eta
		}
	}
	if o.Delivery != nil {
		d := *o.Delivery
		if o.Delivery.ETA != nil {
			eta := *o.Delivery.ETA
			d.ETA = &eta
		}
		c.Delivery = &d
	}
	if o.Payment != nil {
		p := *o.Payment
		c.Payment = &p
	}
	if o.PaidAt != nil {
		t := *o.PaidAt
		c.PaidAt = &t
	}
	return c
}

// IsEmpty reports whether the cart has no units to buy (lines set aside at
// zero do not count).
func (o *Order) IsEmpty() bool {
	for _, it := range o.Items {
		if it.Quantity > 0 {
			return false
		}
	}
	return true
}

// Subtotal sums the items.
func (o *Order) Subtotal() Money {
	var t Money
	for _, it := range o.Items {
		t += it.Subtotal()
	}
	return t
}

// Total adds the delivery fee when the order is delivered at home.
func (o *Order) Total() Money {
	t := o.Subtotal()
	if o.Mode == ModeDelivery && !o.IsEmpty() {
		t += o.DeliveryFee
	}
	return t
}

// IsPaid reports whether the money was collected (any post-payment status).
func (o *Order) IsPaid() bool {
	switch o.Status {
	case OrderPaid, OrderPreparing, OrderReady, OrderDispatched, OrderDelivered:
		return true
	}
	return false
}

// ReservationActive: units are held only while pending and before expiry.
func (o *Order) ReservationActive(now time.Time) bool {
	return o.Status == OrderPending && now.Before(o.ReservationExpiresAt)
}

// SecondsLeft of the reservation, never negative.
func (o *Order) SecondsLeft(now time.Time) int {
	if !o.ReservationActive(now) {
		return 0
	}
	return int(math.Ceil(o.ReservationExpiresAt.Sub(now).Seconds()))
}

// EnsureEditable: the cart can change only while the reservation holds.
func (o *Order) EnsureEditable(now time.Time) error {
	switch {
	case o.Status == OrderExpired, o.Status == OrderPending && !now.Before(o.ReservationExpiresAt):
		return NewError(CodeReservationExpired, "La reserva del pedido %s venció", o.Code)
	case o.Status != OrderPending:
		return NewError(CodeInvalidState, "El pedido %s ya no admite cambios (estado %s)", o.Code, o.Status)
	}
	return nil
}

// EnsurePayable adds the non-empty rule on top of EnsureEditable.
func (o *Order) EnsurePayable(now time.Time) error {
	if err := o.EnsureEditable(now); err != nil {
		return err
	}
	if o.IsEmpty() {
		return NewError(CodeEmptyCart, "El carrito está vacío; no puede pagarse")
	}
	return nil
}

// Item finds a cart line by id.
func (o *Order) Item(itemID string) (OrderItem, error) {
	for _, it := range o.Items {
		if it.ID == itemID {
			return it, nil
		}
	}
	return OrderItem{}, NotFound("Producto del carrito", itemID)
}

// QuantityChange validates the web-page cart rules and returns the delta in
// sale units to reserve (positive) or release (negative). It does not mutate.
func (o *Order) QuantityChange(itemID string, qty int) (OrderItem, int, error) {
	item, err := o.Item(itemID)
	if err != nil {
		return OrderItem{}, 0, err
	}
	if qty < 0 {
		return item, 0, NewError(CodeValidation, "La cantidad no puede ser negativa")
	}
	delta := qty - item.Quantity
	if delta > 0 && item.RequiresPrescription && qty > item.PrescribedQuantity {
		return item, 0, NewError(CodeRxIncreaseNotAllowed,
			"%s es un producto bajo receta: la cantidad máxima es la prescrita (%d %s)", item.Brand, item.PrescribedQuantity, item.UnitLabel)
	}
	return item, delta, nil
}

// ApplyQuantity sets the new quantity. A line at zero stays in the cart so the
// client can add it back; MarkPaid drops it.
func (o *Order) ApplyQuantity(itemID string, qty int) {
	for i := range o.Items {
		if o.Items[i].ID == itemID {
			o.Items[i].Quantity = qty
			return
		}
	}
}

// ItemsFor lists the lines one pharmacy must prepare.
func (o *Order) ItemsFor(pharmacyID string) []OrderItem {
	var out []OrderItem
	for _, it := range o.Items {
		if it.PharmacyID == pharmacyID {
			out = append(out, it)
		}
	}
	return out
}

// PharmacyIDs in first-seen order.
func (o *Order) PharmacyIDs() []string {
	seen := map[string]bool{}
	var ids []string
	for _, it := range o.Items {
		if !seen[it.PharmacyID] {
			seen[it.PharmacyID] = true
			ids = append(ids, it.PharmacyID)
		}
	}
	return ids
}

// FulfillmentFor returns the preparation record of one pharmacy.
func (o *Order) FulfillmentFor(pharmacyID string) *Fulfillment {
	for i := range o.Fulfillments {
		if o.Fulfillments[i].PharmacyID == pharmacyID {
			return &o.Fulfillments[i]
		}
	}
	return nil
}

// MarkPaid moves the order to PAID, drops the lines left at zero and opens the preparation records with the
// demo ETAs. Calling it on a paid order is a no-op (idempotent payments).
func (o *Order) MarkPaid(now time.Time) {
	if o.IsPaid() {
		return
	}
	o.Status = OrderPaid
	paid := now
	o.PaidAt = &paid
	bought := make([]OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		if it.Quantity > 0 {
			bought = append(bought, it)
		}
	}
	o.Items = bought
	o.Fulfillments = nil
	for _, id := range o.PharmacyIDs() {
		items := o.ItemsFor(id)
		f := Fulfillment{PharmacyID: id, PharmacyName: items[0].PharmacyName, Address: items[0].PharmacyAddress,
			Status: FulfillmentNotified}
		if o.Mode == ModePickup {
			eta := now.Add(PickupETA)
			f.ETA = &eta
			f.ETAMinutes = int(PickupETA.Minutes())
		}
		o.Fulfillments = append(o.Fulfillments, f)
	}
	if o.Mode == ModeDelivery {
		if o.Delivery == nil {
			o.Delivery = &Delivery{}
		}
		eta := now.Add(DeliveryETA)
		o.Delivery.Status = DeliveryConsolidating
		o.Delivery.ETA = &eta
		o.Delivery.ETAMinutes = int(DeliveryETA.Minutes())
	}
}

var fulfillmentRank = map[FulfillmentStatus]int{
	FulfillmentNotified: 0, FulfillmentPreparing: 1, FulfillmentReady: 2, FulfillmentPickedUp: 3,
}

// AdvancePharmacy moves one pharmacy's preparation forward and re-derives the
// order status for pickups.
func (o *Order) AdvancePharmacy(pharmacyID string, next FulfillmentStatus, now time.Time) error {
	if !o.IsPaid() {
		return NewError(CodeInvalidState, "El pedido %s aún no está pagado", o.Code)
	}
	f := o.FulfillmentFor(pharmacyID)
	if f == nil {
		return NotFound("Farmacia del pedido", pharmacyID)
	}
	nextRank, ok := fulfillmentRank[next]
	if !ok {
		return NewError(CodeValidation, "Estado de preparación inválido: %s", next)
	}
	if nextRank <= fulfillmentRank[f.Status] {
		return NewError(CodeInvalidState, "La farmacia %s ya está en estado %s", f.PharmacyName, f.Status)
	}
	f.Status = next
	if next == FulfillmentReady || next == FulfillmentPickedUp {
		t := now
		f.ETA = &t
		f.ETAMinutes = 0
	}
	if o.Mode == ModePickup {
		o.Status = o.pickupStatus()
	}
	return nil
}

func (o *Order) pickupStatus() OrderStatus {
	allPicked, allReady, anyPreparing := true, true, false
	for _, f := range o.Fulfillments {
		r := fulfillmentRank[f.Status]
		allPicked = allPicked && r >= fulfillmentRank[FulfillmentPickedUp]
		allReady = allReady && r >= fulfillmentRank[FulfillmentReady]
		anyPreparing = anyPreparing || r >= fulfillmentRank[FulfillmentPreparing]
	}
	switch {
	case allPicked:
		return OrderDelivered
	case allReady:
		return OrderReady
	case anyPreparing:
		return OrderPreparing
	}
	return OrderPaid
}

// AdvanceDelivery moves a home delivery forward.
func (o *Order) AdvanceDelivery(next DeliveryStatus, now time.Time) error {
	if o.Mode != ModeDelivery || o.Delivery == nil {
		return NewError(CodeInvalidState, "El pedido %s no es a domicilio", o.Code)
	}
	if !o.IsPaid() {
		return NewError(CodeInvalidState, "El pedido %s aún no está pagado", o.Code)
	}
	switch {
	case next == DeliveryDispatched && o.Delivery.Status == DeliveryConsolidating:
		eta := now.Add(DispatchETA)
		o.Delivery.Status = DeliveryDispatched
		o.Delivery.ETA = &eta
		o.Delivery.ETAMinutes = int(DispatchETA.Minutes())
		o.Status = OrderDispatched
	case next == DeliveryDelivered && o.Delivery.Status == DeliveryDispatched:
		t := now
		o.Delivery.Status = DeliveryDelivered
		o.Delivery.ETA = &t
		o.Delivery.ETAMinutes = 0
		o.Status = OrderDelivered
	default:
		return NewError(CodeInvalidState, "Transición de entrega inválida: %s → %s", o.Delivery.Status, next)
	}
	return nil
}

// Cancel is allowed only before the money is collected.
func (o *Order) Cancel() error {
	if o.Status != OrderPending && o.Status != OrderExpired {
		return NewError(CodeInvalidState, "El pedido %s no puede cancelarse (estado %s)", o.Code, o.Status)
	}
	o.Status = OrderCancelled
	return nil
}

// Expire marks a pending order whose reservation ran out.
func (o *Order) Expire() {
	if o.Status == OrderPending {
		o.Status = OrderExpired
	}
}

// Renew re-opens an expired order with a fresh reservation window.
func (o *Order) Renew(now time.Time, ttl time.Duration) error {
	if o.Status != OrderExpired {
		return NewError(CodeInvalidState, "El pedido %s no está vencido (estado %s)", o.Code, o.Status)
	}
	o.Status = OrderPending
	o.ReservationExpiresAt = now.Add(ttl)
	o.Payment = nil
	return nil
}

// Describe is a short human label used in notifications: "DEMO-001".
func (o *Order) Describe() string { return fmt.Sprintf("%s", o.Code) }
