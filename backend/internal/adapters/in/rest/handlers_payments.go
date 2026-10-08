package rest

import (
	"net/http"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

func (h *handlers) createPayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderID string `json:"orderId"`
		Method  string `json:"method"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.OrderID == "" {
		writeError(w, domain.NewError(domain.CodeValidation, "orderId es obligatorio"))
		return
	}
	p, err := h.s.Payments.Create(r.Context(), req.OrderID, domain.PaymentMethod(req.Method))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"payment": toPaymentDTO(p)})
}

func (h *handlers) getPayment(w http.ResponseWriter, r *http.Request) {
	p, o, err := h.s.Payments.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"payment": toPaymentDTO(p),
		"order": orderBriefDTO{ID: o.ID, Code: o.Code, Total: o.Total().Float(), Status: string(o.Status)}})
}

func (h *handlers) banks(w http.ResponseWriter, r *http.Request) {
	banks, err := h.s.Payments.Banks(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]bankDTO, 0, len(banks))
	for _, b := range banks {
		out = append(out, bankDTO(b))
	}
	writeJSON(w, http.StatusOK, map[string]any{"banks": out})
}

func (h *handlers) confirmPayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Outcome string `json:"outcome"`
		BankID  string `json:"bankId"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	p, o, err := h.s.Payments.Confirm(r.Context(), r.PathValue("id"), domain.PaymentOutcome(req.Outcome), req.BankID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"payment": toPaymentDTO(p), "order": toOrderDTO(o, h.s.Clock.Now())})
}
