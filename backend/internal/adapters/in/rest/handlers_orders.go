package rest

import (
	"net/http"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

func (h *handlers) respondOrder(w http.ResponseWriter, o domain.Order, err error) {
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"order": toOrderDTO(o, h.s.Clock.Now())})
}

func (h *handlers) getOrder(w http.ResponseWriter, r *http.Request) {
	o, err := h.s.Orders.Get(r.Context(), r.PathValue("id"))
	h.respondOrder(w, o, err)
}

func (h *handlers) changeQuantity(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Quantity *int `json:"quantity"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.Quantity == nil {
		writeError(w, domain.NewError(domain.CodeValidation, "quantity es obligatorio"))
		return
	}
	o, err := h.s.Orders.ChangeQuantity(r.Context(), r.PathValue("id"), r.PathValue("itemId"), *req.Quantity)
	h.respondOrder(w, o, err)
}

func (h *handlers) cancelOrder(w http.ResponseWriter, r *http.Request) {
	o, err := h.s.Orders.Cancel(r.Context(), r.PathValue("id"))
	h.respondOrder(w, o, err)
}

func (h *handlers) renewOrder(w http.ResponseWriter, r *http.Request) {
	o, err := h.s.Orders.Renew(r.Context(), r.PathValue("id"))
	h.respondOrder(w, o, err)
}

func (h *handlers) paymentOptions(w http.ResponseWriter, r *http.Request) {
	opts, err := h.s.Orders.PaymentOptions(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]paymentOptionDTO, 0, len(opts.Options))
	for _, o := range opts.Options {
		out = append(out, paymentOptionDTO{Method: string(o.Method), Label: o.Label, Description: o.Description})
	}
	writeJSON(w, http.StatusOK, map[string]any{"checkoutUrl": opts.CheckoutURL, "options": out})
}
