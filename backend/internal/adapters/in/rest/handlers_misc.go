package rest

import (
	"net/http"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

func (h *handlers) notifications(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := h.s.Notifications.List(r.Context(), domain.NotificationFilter{Channel: domain.Channel(q.Get("channel")), OrderID: q.Get("orderId")})
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]notificationDTO, 0, len(list))
	for _, n := range list {
		out = append(out, notificationDTO{ID: n.ID, Channel: string(n.Channel), Recipient: n.Recipient, Title: n.Title,
			Body: n.Body, OrderID: n.OrderID, At: ts(n.At)})
	}
	writeJSON(w, http.StatusOK, map[string]any{"notifications": out})
}

func (h *handlers) getClient(w http.ResponseWriter, r *http.Request) {
	c, err := h.s.Clients.Find(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"client": clientDTO(c)})
}

func (h *handlers) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "surface": h.cfg.Surface, "modules": h.cfg.Modules})
}

func (h *handlers) apiRoot(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"service": "farmi-backend", "surface": h.cfg.Surface, "api": "/api/v1", "docs": "docs/API.md"})
}

func (h *handlers) apiNotFound(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusNotFound, errorEnvelope{errorBody{"NOT_FOUND", "Ruta no encontrada"}})
}
