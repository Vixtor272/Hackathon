// Package farmi is the WhatsApp sales assistant: a state machine over the
// conversation that orchestrates OCR, validation, availability, orders and
// payments exclusively through ports.
package farmi

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Deps groups the collaborators of the assistant.
type Deps struct {
	Conversations ports.ConversationRepository
	Log           ports.MessageLog
	Sender        ports.MessageSender
	Clients       ports.ClientRepository
	OCR           ports.OCRService
	Validator     ports.PrescriptionValidator
	Availability  ports.AvailabilityService
	Orders        ports.OrderService
	AI            ports.ConversationAI
	Clock         ports.Clock
	IDs           ports.IDGenerator
	Courier       domain.Courier
}

// Assistant implements ports.Assistant.
type Assistant struct {
	d  Deps
	mu sync.Mutex // one message at a time keeps the per-phone state consistent
}

// New builds Farmi.
func New(d Deps) *Assistant { return &Assistant{d: d} }

// turn is everything about handling one inbound message.
type turn struct {
	conv   *domain.Conversation
	in     ports.InboundMessage
	intent domain.Interpretation
	fresh  bool
	out    []domain.Message
}

func (t *turn) say(text string) {
	t.out = append(t.out, domain.Message{Direction: domain.DirectionOut, Type: domain.MessageText, Text: text})
}

func (t *turn) link(text, url string) {
	t.out = append(t.out, domain.Message{Direction: domain.DirectionOut, Type: domain.MessageLink, Text: text, Link: url})
}

func (t *turn) isImage() bool { return t.in.Type == domain.MessageImage }

func (t *turn) choiceWithin(n int) (int, bool) {
	if t.intent.Intent != domain.IntentChoice || t.intent.Choice < 1 || t.intent.Choice > n {
		return 0, false
	}
	return t.intent.Choice, true
}

// HandleInbound processes a client message and sends Farmi's replies.
func (a *Assistant) HandleInbound(ctx context.Context, in ports.InboundMessage) (ports.ConversationReply, error) {
	phone := domain.NormalizePhone(in.From)
	if phone == "" {
		return ports.ConversationReply{}, domain.NewError(domain.CodeValidation, "Falta el número de teléfono")
	}
	if in.Type == "" {
		in.Type = domain.MessageText
	}
	in.From = phone

	a.mu.Lock()
	defer a.mu.Unlock()

	conv, found, err := a.d.Conversations.Find(ctx, phone)
	if err != nil {
		return ports.ConversationReply{}, err
	}
	if !found {
		conv = domain.NewConversation(phone)
	}
	now := a.d.Clock.Now()
	inbound := domain.Message{ID: a.d.IDs.New("msg"), Direction: domain.DirectionIn, Type: in.Type,
		Text: in.Text, MediaID: in.MediaID, At: now}
	if in.Type == domain.MessageImage {
		inbound.MediaURL = a.mediaURL(ctx, in.MediaID)
	}
	if err := a.d.Log.Append(ctx, phone, inbound); err != nil {
		return ports.ConversationReply{}, err
	}

	t := &turn{conv: &conv, in: in, fresh: !found}
	if in.Type != domain.MessageImage {
		t.intent = a.d.AI.Interpret(ctx, in.Text)
	}
	handleErr := a.dispatch(ctx, t)
	if handleErr != nil {
		t.say(msgInternalError)
	}
	if n := len(t.out); n > 0 {
		t.out[n-1].Options = a.quickReplies(ctx, &conv)
	}
	conv.UpdatedAt = now
	if err := a.d.Conversations.Save(ctx, conv); err != nil {
		return ports.ConversationReply{}, err
	}
	for i := range t.out {
		t.out[i].ID = a.d.IDs.New("msg")
		t.out[i].At = now
		if err := a.d.Sender.Send(ctx, phone, t.out[i]); err != nil {
			return ports.ConversationReply{}, err
		}
	}
	return ports.ConversationReply{State: conv.State, Replies: t.out}, handleErr
}

// Transcript returns the chat with one phone.
func (a *Assistant) Transcript(ctx context.Context, phone string) (ports.Transcript, error) {
	phone = domain.NormalizePhone(phone)
	msgs, err := a.d.Log.List(ctx, phone)
	if err != nil {
		return ports.Transcript{}, err
	}
	tr := ports.Transcript{Phone: phone, State: domain.StateAskID, Messages: msgs}
	if conv, found, err := a.d.Conversations.Find(ctx, phone); err != nil {
		return ports.Transcript{}, err
	} else if found {
		tr.State, tr.OrderID = conv.State, conv.OrderID
	}
	return tr, nil
}

// Reset forgets the conversation and its transcript.
func (a *Assistant) Reset(ctx context.Context, phone string) error {
	phone = domain.NormalizePhone(phone)
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.d.Conversations.Delete(ctx, phone); err != nil {
		return err
	}
	return a.d.Log.Clear(ctx, phone)
}

// SampleMedia lists the prescription images the simulated phone can send.
func (a *Assistant) SampleMedia(ctx context.Context) ([]domain.SampleMedia, error) {
	return a.d.OCR.SampleMedia(ctx)
}

func (a *Assistant) mediaURL(ctx context.Context, mediaID string) string {
	media, err := a.d.OCR.SampleMedia(ctx)
	if err != nil {
		return ""
	}
	for _, m := range media {
		if m.ID == mediaID {
			return m.URL
		}
	}
	return ""
}

// dispatch applies the global commands, then the handler of the current step.
func (a *Assistant) dispatch(ctx context.Context, t *turn) error {
	switch t.intent.Intent {
	case domain.IntentRestart:
		t.conv.StartOver()
		t.say(msgGreeting)
		return nil
	case domain.IntentHelp:
		t.say(helpFor(t.conv.State))
		return nil
	}
	switch t.conv.State {
	case domain.StateAskID:
		return a.askID(ctx, t)
	case domain.StateAskName:
		return a.askName(ctx, t)
	case domain.StateAskPrescription:
		return a.askPrescription(ctx, t)
	case domain.StateAskZone:
		return a.askZone(ctx, t)
	case domain.StateAskMode:
		return a.askMode(ctx, t)
	case domain.StateAskPickupOption:
		return a.askPickupOption(ctx, t)
	case domain.StateAskAddress:
		return a.askAddress(ctx, t)
	case domain.StateAskBrand:
		return a.askBrand(ctx, t)
	case domain.StateConfirmCart:
		return a.confirmCart(ctx, t)
	case domain.StateAwaitPayment:
		return a.awaitPayment(ctx, t)
	case domain.StateCompleted:
		return a.completed(ctx, t)
	}
	t.conv.StartOver()
	t.say(msgGreeting)
	return nil
}

// askID: the purchase never starts without the client's cédula.
func (a *Assistant) askID(ctx context.Context, t *turn) error {
	if t.isImage() {
		t.say(msgIDBeforePrescription)
		return nil
	}
	id := domain.NormalizeCedula(t.in.Text)
	if !domain.ValidCedula(id) {
		if t.fresh || t.intent.Intent == domain.IntentGreeting {
			t.say(msgGreeting)
		} else {
			t.say(msgInvalidID)
		}
		return nil
	}
	client, found, err := a.d.Clients.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !found {
		t.conv.ClientID = id
		t.conv.State = domain.StateAskName
		t.say(msgAskName(id))
		return nil
	}
	a.identify(t, client)
	return nil
}

func (a *Assistant) identify(t *turn, c domain.Client) {
	t.conv.ClientID, t.conv.ClientName = c.ID, c.Name
	t.conv.State = domain.StateAskPrescription
	t.say(msgAskPrescription(firstName(c.Name)))
}

func (a *Assistant) askName(ctx context.Context, t *turn) error {
	name := strings.TrimSpace(t.in.Text)
	if t.isImage() || len(name) < 2 {
		t.say(msgAskNameAgain)
		return nil
	}
	c := domain.Client{ID: t.conv.ClientID, Name: name, Phone: t.conv.Phone}
	if err := a.d.Clients.Save(ctx, c); err != nil {
		return err
	}
	a.identify(t, c)
	return nil
}

// askPrescription: OCR → JSON → validation (fields + doctor registry).
func (a *Assistant) askPrescription(ctx context.Context, t *turn) error {
	if !t.isImage() {
		t.say(msgNeedPhoto)
		return nil
	}
	rx, err := a.d.OCR.Extract(ctx, t.in.MediaID)
	if err != nil {
		if domain.IsCode(err, domain.CodeOCRIllegible) || domain.IsCode(err, domain.CodeNotFound) {
			t.say(msgIllegible)
			return nil
		}
		return err
	}
	res, err := a.d.Validator.Validate(ctx, rx)
	if err != nil {
		return err
	}
	if !res.Valid {
		t.say(msgInvalidPrescription(res.Errors))
		return nil
	}
	t.conv.PrescriptionID = rx.ID
	t.say(msgPrescriptionOK(rx, res))
	return a.promptZone(ctx, t)
}

func (a *Assistant) promptZone(ctx context.Context, t *turn) error {
	zones, err := a.d.Availability.Zones(ctx)
	if err != nil {
		return err
	}
	t.conv.State = domain.StateAskZone
	t.conv.Availability = nil
	t.say(msgAskZone(zones))
	return nil
}

// askZone: consult the company's API for options in the chosen zone. A zone
// with no pharmacy for the prescription keeps the client here to pick another
// one (or to switch to delivery when the company can complete it).
func (a *Assistant) askZone(ctx context.Context, t *turn) error {
	zones, err := a.d.Availability.Zones(ctx)
	if err != nil {
		return err
	}
	if prev := t.conv.Availability; prev != nil && prev.DeliveryAvailable && !t.isImage() && t.intent.Intent == domain.IntentDelivery {
		t.conv.Mode = domain.ModeDelivery
		t.conv.State = domain.StateAskAddress
		t.say(msgAskAddress)
		return nil
	}
	zone, ok := a.d.AI.MatchZone(ctx, t.in.Text, zones)
	if t.isImage() || !ok {
		t.say(msgZoneNotUnderstood(zones))
		return nil
	}
	rx, err := a.d.OCR.Find(ctx, t.conv.PrescriptionID)
	if err != nil {
		return err
	}
	av, err := a.d.Availability.FindOptions(ctx, zone.ID, rx)
	if err != nil {
		return err
	}
	t.conv.Zone = zone
	t.conv.Availability = &av
	if len(av.Options) == 0 {
		t.say(msgNoPharmacies(zone, av.Missing))
		t.say(msgAskOtherZone(zones, av.DeliveryAvailable))
		return nil
	}
	t.say(msgOptions(av))
	t.conv.State = domain.StateAskMode
	t.say(msgAskMode)
	return nil
}

func (a *Assistant) askMode(ctx context.Context, t *turn) error {
	av := t.conv.Availability
	switch t.intent.Intent {
	case domain.IntentPickup:
		if len(av.Options) == 0 {
			t.say(msgNoPickup)
			return nil
		}
		t.conv.Mode = domain.ModePickup
		if len(av.Options) == 1 {
			// Only one way to pick it up: no need to ask which.
			return a.choosePickupOption(ctx, t, av.Options[0])
		}
		t.conv.State = domain.StateAskPickupOption
		t.say(msgAskPickupOption(len(av.Options)))
	case domain.IntentDelivery:
		if !av.DeliveryAvailable {
			t.say(msgNoDelivery)
			return nil
		}
		t.conv.Mode = domain.ModeDelivery
		t.conv.State = domain.StateAskAddress
		t.say(msgAskAddress)
	default:
		t.say(msgModeNotUnderstood)
	}
	return nil
}

func (a *Assistant) askPickupOption(ctx context.Context, t *turn) error {
	av := t.conv.Availability
	n, ok := t.choiceWithin(len(av.Options))
	if !ok {
		t.say(msgAskPickupOption(len(av.Options)))
		return nil
	}
	return a.choosePickupOption(ctx, t, av.Options[n-1])
}

func (a *Assistant) choosePickupOption(ctx context.Context, t *turn, opt domain.PharmacyOption) error {
	t.conv.Coverage = opt.Coverage
	t.say(msgOptionChosen(opt))
	return a.startBrands(ctx, t)
}

// askAddress: delivery uses company-wide stock; origins are chosen internally.
func (a *Assistant) askAddress(ctx context.Context, t *turn) error {
	addr := strings.TrimSpace(t.in.Text)
	if t.isImage() || len(addr) < 5 {
		t.say(msgAskAddressAgain)
		return nil
	}
	rx, err := a.d.OCR.Find(ctx, t.conv.PrescriptionID)
	if err != nil {
		return err
	}
	plan, err := a.d.Availability.PlanDelivery(ctx, rx)
	if err != nil {
		return err
	}
	if !plan.Feasible() {
		t.say(msgNoDeliveryNow(plan.Missing))
		t.conv.State = domain.StateAskMode
		t.say(msgAskMode)
		return nil
	}
	courier := a.d.Courier
	t.conv.DeliveryAddress = addr
	t.conv.Coverage = plan.Coverage
	t.conv.Courier = &courier
	t.say(msgDeliveryAssigned(addr, courier, domain.DefaultDeliveryFee))
	return a.startBrands(ctx, t)
}

func (a *Assistant) startBrands(ctx context.Context, t *turn) error {
	rx, err := a.d.OCR.Find(ctx, t.conv.PrescriptionID)
	if err != nil {
		return err
	}
	brands, err := a.d.Availability.BrandOptions(ctx, t.conv.Coverage, rx)
	if err != nil {
		return err
	}
	for _, mb := range brands {
		if len(mb.Brands) == 0 {
			t.say(msgNoBrands(mb.Medicine))
			return a.promptZone(ctx, t)
		}
	}
	if len(brands) == 0 {
		return a.promptZone(ctx, t)
	}
	t.conv.ResetBrands(brands)
	t.conv.State = domain.StateAskBrand
	t.say(msgBrandPrompt(brands[0]))
	return nil
}

func (a *Assistant) askBrand(ctx context.Context, t *turn) error {
	current, ok := t.conv.CurrentBrands()
	if !ok {
		return a.startBrands(ctx, t)
	}
	n, ok := t.choiceWithin(len(current.Brands))
	if !ok {
		t.say(msgChooseNumber(len(current.Brands)))
		t.say(msgBrandPrompt(current))
		return nil
	}
	t.conv.Select(current.Brands[n-1])
	if next, ok := t.conv.CurrentBrands(); ok {
		t.say(msgBrandPrompt(next))
		return nil
	}
	t.conv.State = domain.StateConfirmCart
	t.say(msgCart(t.conv))
	return nil
}

func (a *Assistant) confirmCart(ctx context.Context, t *turn) error {
	switch t.intent.Intent {
	case domain.IntentYes:
		return a.placeOrder(ctx, t)
	case domain.IntentNo:
		t.say(msgRechooseBrands)
		return a.startBrands(ctx, t)
	}
	t.say(msgConfirmNotUnderstood)
	return nil
}

// placeOrder re-checks stock, reserves for 10 minutes and sends the link.
func (a *Assistant) placeOrder(ctx context.Context, t *turn) error {
	in := ports.CreateOrderInput{
		Phone: t.conv.Phone, ClientID: t.conv.ClientID, ClientName: t.conv.ClientName,
		PrescriptionID: t.conv.PrescriptionID, Mode: t.conv.Mode, Zone: t.conv.Zone,
		DeliveryAddress: t.conv.DeliveryAddress, Courier: t.conv.Courier, Selections: t.conv.Selections,
	}
	order, err := a.d.Orders.Create(ctx, in)
	if err != nil {
		if domain.IsCode(err, domain.CodeInsufficientStock) {
			t.say(msgReservationFailed(messageOf(err)))
			return a.startBrands(ctx, t)
		}
		return err
	}
	opts, err := a.d.Orders.PaymentOptions(ctx, order.ID)
	if err != nil {
		return err
	}
	t.conv.OrderID = order.ID
	t.conv.State = domain.StateAwaitPayment
	t.say(msgReserved(order, a.d.Clock.Now()))
	t.link(msgPayLink(opts), opts.CheckoutURL)
	return nil
}

// awaitPayment answers status questions and handles cancel / retry while the
// web page collects the payment.
func (a *Assistant) awaitPayment(ctx context.Context, t *turn) error {
	order, err := a.d.Orders.Get(ctx, t.conv.OrderID)
	if err != nil {
		return err
	}
	now := a.d.Clock.Now()
	if order.IsPaid() {
		t.conv.State = domain.StateCompleted
		t.say(msgStatus(order, now))
		return nil
	}
	switch t.intent.Intent {
	case domain.IntentStatus:
		t.say(msgStatus(order, now))
		if order.Status == domain.OrderExpired {
			t.say(msgExpiredHint)
		}
	case domain.IntentCancel:
		if _, err := a.d.Orders.Cancel(ctx, order.ID); err != nil {
			if _, ok := domain.CodeOf(err); ok {
				t.say(messageOf(err))
				return nil
			}
			return err
		}
		t.conv.State = domain.StateCompleted
		t.say(msgAfterCancel)
	case domain.IntentContinue, domain.IntentPayLink, domain.IntentYes:
		return a.resumePayment(ctx, t, order, now)
	default:
		t.say(msgStatus(order, now))
		switch order.Status {
		case domain.OrderPending:
			opts, err := a.d.Orders.PaymentOptions(ctx, order.ID)
			if err != nil {
				return err
			}
			t.link(msgPayLink(opts), opts.CheckoutURL)
		case domain.OrderExpired:
			t.say(msgExpiredHint)
		default:
			t.conv.State = domain.StateCompleted
			t.say(msgNewPurchaseHint)
		}
	}
	return nil
}

// resumePayment re-sends the link, renewing the reservation first if it lapsed.
func (a *Assistant) resumePayment(ctx context.Context, t *turn, order domain.Order, now time.Time) error {
	if order.Status == domain.OrderCancelled {
		t.conv.State = domain.StateCompleted
		t.say(msgStatus(order, now))
		t.say(msgNewPurchaseHint)
		return nil
	}
	if !order.ReservationActive(now) {
		renewed, err := a.d.Orders.Renew(ctx, order.ID)
		if err != nil {
			if _, ok := domain.CodeOf(err); ok {
				t.say(msgRenewFailed(messageOf(err)))
				t.conv.State = domain.StateCompleted
				t.say(msgNewPurchaseHint)
				return nil
			}
			return err
		}
		order = renewed
		t.say(msgRenewed(order))
	}
	opts, err := a.d.Orders.PaymentOptions(ctx, order.ID)
	if err != nil {
		return err
	}
	t.link(msgPayLink(opts), opts.CheckoutURL)
	return nil
}

// completed: after a purchase any new message starts over, and Farmi asks
// for the cédula again before the next one.
func (a *Assistant) completed(ctx context.Context, t *turn) error {
	if t.intent.Intent == domain.IntentStatus && t.conv.OrderID != "" {
		order, err := a.d.Orders.Get(ctx, t.conv.OrderID)
		if err != nil {
			return err
		}
		t.say(msgStatus(order, a.d.Clock.Now()))
		return nil
	}
	t.conv.StartOver()
	t.say(msgGreetingAgain)
	return nil
}

func firstName(name string) string {
	if parts := strings.Fields(name); len(parts) > 0 {
		return parts[0]
	}
	return name
}

func messageOf(err error) string {
	var e *domain.Error
	if errors.As(err, &e) {
		return e.Message
	}
	return err.Error()
}
