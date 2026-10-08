// Package simulator is the payment rail of the demo: no real processor, no
// bank data. DeUna attempts get an internal link; cards are confirmed inline.
package simulator

import (
	"context"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

// Gateway implements ports.PaymentGateway.
type Gateway struct {
	webBaseURL string
	banks      []domain.Bank
}

// NewGateway builds the simulator; links point at the Svelte checkout.
func NewGateway(webBaseURL string, banks []domain.Bank) *Gateway {
	return &Gateway{webBaseURL: webBaseURL, banks: banks}
}

// CreateIntent attaches the simulator link for DeUna attempts.
func (g *Gateway) CreateIntent(_ context.Context, p domain.Payment) (domain.Payment, error) {
	if p.Method == domain.MethodDeUna {
		p.Link = g.webBaseURL + "/deuna/" + p.ID
	}
	return p, nil
}

// ListBanks returns the fictitious banks.
func (g *Gateway) ListBanks(_ context.Context) ([]domain.Bank, error) {
	return append([]domain.Bank(nil), g.banks...), nil
}
