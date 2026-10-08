package farmi

import (
	"context"
	"fmt"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

var (
	optPickup      = domain.MessageOption{Label: "🏪 Retiro", Value: "retiro"}
	optDelivery    = domain.MessageOption{Label: "🚚 Domicilio", Value: "domicilio"}
	optYes         = domain.MessageOption{Label: "✅ Sí, confirmar", Value: "sí"}
	optNo          = domain.MessageOption{Label: "🔁 No, cambiar marcas", Value: "no"}
	optStatus      = domain.MessageOption{Label: "📋 Estado", Value: "estado"}
	optCancel      = domain.MessageOption{Label: "❌ Cancelar pedido", Value: "cancelar"}
	optContinue    = domain.MessageOption{Label: "🔄 Continuar", Value: "continuar"}
	optNewPurchase = domain.MessageOption{Label: "🛒 Nueva compra", Value: "nueva compra"}
)

func numbered(n int, label string) domain.MessageOption {
	return domain.MessageOption{Label: fmt.Sprintf("%d. %s", n, label), Value: fmt.Sprint(n)}
}

// quickReplies lists the answers that make sense at the step the conversation
// is left in, so the client is offered only those. Steps answered with free
// text or a photo (cédula, name, prescription, address) offer none.
func (a *Assistant) quickReplies(ctx context.Context, conv *domain.Conversation) []domain.MessageOption {
	var opts []domain.MessageOption
	switch conv.State {
	case domain.StateAskZone:
		zones, err := a.d.Availability.Zones(ctx)
		if err != nil {
			return nil
		}
		for i, z := range zones {
			opts = append(opts, numbered(i+1, z.Label))
		}
		if av := conv.Availability; av != nil && av.DeliveryAvailable {
			opts = append(opts, optDelivery)
		}
	case domain.StateAskMode:
		if av := conv.Availability; av != nil {
			if len(av.Options) > 0 {
				opts = append(opts, optPickup)
			}
			if av.DeliveryAvailable {
				opts = append(opts, optDelivery)
			}
		}
	case domain.StateAskPickupOption:
		if av := conv.Availability; av != nil {
			for i, o := range av.Options {
				opts = append(opts, numbered(i+1, o.Label))
			}
		}
	case domain.StateAskBrand:
		if mb, ok := conv.CurrentBrands(); ok {
			for i, br := range mb.Brands {
				opts = append(opts, numbered(i+1, br.Product.Brand))
			}
		}
	case domain.StateConfirmCart:
		opts = []domain.MessageOption{optYes, optNo}
	case domain.StateAwaitPayment:
		order, err := a.d.Orders.Get(ctx, conv.OrderID)
		if err != nil {
			return nil
		}
		if order.ReservationActive(a.d.Clock.Now()) {
			return []domain.MessageOption{optStatus, optCancel}
		}
		opts = []domain.MessageOption{optContinue, optCancel}
	case domain.StateCompleted:
		if order, err := a.d.Orders.Get(ctx, conv.OrderID); err == nil && order.IsPaid() {
			opts = append(opts, optStatus)
		}
		opts = append(opts, optNewPurchase)
	}
	return opts
}
