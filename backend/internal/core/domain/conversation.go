package domain

import "time"

// ConversationState is the step Farmi is at with one phone number.
type ConversationState string

const (
	StateAskID           ConversationState = "ASK_ID"
	StateAskName         ConversationState = "ASK_NAME"
	StateAskPrescription ConversationState = "ASK_PRESCRIPTION"
	StateAskZone         ConversationState = "ASK_ZONE"
	StateAskMode         ConversationState = "ASK_MODE"
	StateAskPickupOption ConversationState = "ASK_PICKUP_OPTION"
	StateAskAddress      ConversationState = "ASK_ADDRESS"
	StateAskBrand        ConversationState = "ASK_BRAND"
	StateConfirmCart     ConversationState = "CONFIRM_CART"
	StateAwaitPayment    ConversationState = "AWAIT_PAYMENT"
	StateCompleted       ConversationState = "COMPLETED"
)

// MessageDirection: "in" comes from the client, "out" from Farmi.
type MessageDirection string

const (
	DirectionIn  MessageDirection = "in"
	DirectionOut MessageDirection = "out"
)

// MessageType is the WhatsApp payload kind the simulator supports.
type MessageType string

const (
	MessageText  MessageType = "text"
	MessageImage MessageType = "image"
	MessageLink  MessageType = "link"
)

// Message is one bubble of the WhatsApp transcript.
type Message struct {
	ID        string
	Direction MessageDirection
	Type      MessageType
	Text      string
	MediaID   string
	MediaURL  string
	Link      string
	At        time.Time
}

// Intent is what the AI module understood from a free-text reply.
type Intent string

const (
	IntentFreeText Intent = "free_text"
	IntentChoice   Intent = "choice"
	IntentGreeting Intent = "greeting"
	IntentYes      Intent = "yes"
	IntentNo       Intent = "no"
	IntentPickup   Intent = "pickup"
	IntentDelivery Intent = "delivery"
	IntentStatus   Intent = "status"
	IntentCancel   Intent = "cancel"
	IntentContinue Intent = "continue"
	IntentRestart  Intent = "restart"
	IntentHelp     Intent = "help"
	IntentPayLink  Intent = "pay_link"
)

// Interpretation is the structured reading of a client message.
type Interpretation struct {
	Intent Intent
	Choice int // 1-based when Intent == IntentChoice
}

// Conversation is the purchase in progress for one phone number. Everything
// Farmi needs between two messages lives here, so the assistant is stateless.
type Conversation struct {
	Phone           string
	State           ConversationState
	ClientID        string
	ClientName      string
	PrescriptionID  string
	Zone            Zone
	Mode            FulfillmentMode
	Availability    *Availability
	Coverage        []Coverage
	DeliveryAddress string
	Courier         *Courier
	Brands          []MedicineBrands
	BrandIndex      int
	Selections      []BrandOption
	OrderID         string
	UpdatedAt       time.Time
}

// NewConversation starts at the identification step: Farmi always asks for
// the cédula before a purchase begins.
func NewConversation(phone string) Conversation {
	return Conversation{Phone: phone, State: StateAskID}
}

// StartOver forgets the purchase in progress but keeps the phone.
func (c *Conversation) StartOver() {
	*c = NewConversation(c.Phone)
}

// ResetBrands restarts the brand selection from the first medicine.
func (c *Conversation) ResetBrands(brands []MedicineBrands) {
	c.Brands = brands
	c.BrandIndex = 0
	c.Selections = nil
}

// CurrentBrands is the medicine whose brand is being chosen.
func (c *Conversation) CurrentBrands() (MedicineBrands, bool) {
	if c.BrandIndex < 0 || c.BrandIndex >= len(c.Brands) {
		return MedicineBrands{}, false
	}
	return c.Brands[c.BrandIndex], true
}

// Select records the chosen brand and advances to the next medicine.
func (c *Conversation) Select(b BrandOption) {
	c.Selections = append(c.Selections, b)
	c.BrandIndex++
}

// BrandsDone reports whether every medicine has a chosen brand.
func (c *Conversation) BrandsDone() bool { return c.BrandIndex >= len(c.Brands) }

// CartSubtotal sums the chosen brands.
func (c *Conversation) CartSubtotal() Money {
	var t Money
	for _, s := range c.Selections {
		t += s.Subtotal()
	}
	return t
}
