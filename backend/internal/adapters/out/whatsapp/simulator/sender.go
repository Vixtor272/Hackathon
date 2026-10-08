// Package simulator replaces the WhatsApp Cloud API: an outbound message is
// "delivered" by appending it to the phone's transcript, which the simulated
// phone in the web UI reads.
package simulator

import (
	"context"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Sender implements ports.MessageSender.
type Sender struct {
	log ports.MessageLog
}

// NewSender wires the transcript store.
func NewSender(log ports.MessageLog) *Sender { return &Sender{log: log} }

// Send delivers the message to the simulated phone.
func (s *Sender) Send(ctx context.Context, phone string, m domain.Message) error {
	return s.log.Append(ctx, phone, m)
}
