package domain

import "time"

// PaymentMethod is one of the simulated payment rails.
type PaymentMethod string

const (
	MethodCard  PaymentMethod = "card"
	MethodDeUna PaymentMethod = "deuna"
)

// PaymentStatus is the lifecycle of one payment attempt.
type PaymentStatus string

const (
	PaymentPending     PaymentStatus = "PENDING"
	PaymentApproved    PaymentStatus = "APPROVED"
	PaymentRejected    PaymentStatus = "REJECTED"
	PaymentInvalidated PaymentStatus = "INVALIDATED"
)

// PaymentOutcome is what the simulator button reports.
type PaymentOutcome string

const (
	OutcomeApproved PaymentOutcome = "approved"
	OutcomeRejected PaymentOutcome = "rejected"
)

// Payment is one attempt to pay an order for a fixed amount.
type Payment struct {
	ID          string
	OrderID     string
	Method      PaymentMethod
	Amount      Money
	Status      PaymentStatus
	Link        string
	BankID      string
	CreatedAt   time.Time
	ConfirmedAt *time.Time
}

// Ref is the compact view stored on the order.
func (p Payment) Ref() PaymentRef { return PaymentRef{ID: p.ID, Method: p.Method, Status: p.Status} }

// PaymentOption is a rail offered to the client.
type PaymentOption struct {
	Method      PaymentMethod
	Label       string
	Description string
}

// PaymentOptions lists the rails of the demo.
func PaymentOptions() []PaymentOption {
	return []PaymentOption{
		{Method: MethodCard, Label: "Tarjeta de débito o crédito", Description: "Simulador sin datos bancarios reales"},
		{Method: MethodDeUna, Label: "DeUna", Description: "Paga con tu banco (simulado)"},
	}
}

// ValidPaymentMethod guards user input.
func ValidPaymentMethod(m PaymentMethod) bool { return m == MethodCard || m == MethodDeUna }

// Bank is a fictitious bank shown by the DeUna simulator.
type Bank struct {
	ID   string
	Name string
}
