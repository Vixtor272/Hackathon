package rest

import (
	"net/http"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

func (h *handlers) fulfillmentOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	orders, err := h.s.Fulfillment.List(r.Context(), ports.FulfillmentFilter{PharmacyID: q.Get("pharmacyId"), Role: q.Get("role")})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": toOrderDTOs(orders, h.s.Clock.Now())})
}

func (h *handlers) pharmacyStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Status string `json:"status"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	o, err := h.s.Fulfillment.AdvancePharmacy(r.Context(), r.PathValue("id"), r.PathValue("pharmacyId"), domain.FulfillmentStatus(req.Status))
	h.respondOrder(w, o, err)
}

func (h *handlers) deliveryStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Status string `json:"status"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	o, err := h.s.Fulfillment.AdvanceDelivery(r.Context(), r.PathValue("id"), domain.DeliveryStatus(req.Status))
	h.respondOrder(w, o, err)
}
