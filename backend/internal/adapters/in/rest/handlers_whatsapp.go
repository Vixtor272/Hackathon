package rest

import (
	"net/http"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

type webhookRequest struct {
	From    string `json:"from"`
	Type    string `json:"type"`
	Text    string `json:"text"`
	MediaID string `json:"mediaId"`
}

func (h *handlers) webhook(w http.ResponseWriter, r *http.Request) {
	var req webhookRequest
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	reply, err := h.s.Assistant.HandleInbound(r.Context(), ports.InboundMessage{
		From: req.From, Type: domain.MessageType(req.Type), Text: req.Text, MediaID: req.MediaID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"state": reply.State, "replies": toMessageDTOs(reply.Replies)})
}

func (h *handlers) transcript(w http.ResponseWriter, r *http.Request) {
	tr, err := h.s.Assistant.Transcript(r.Context(), r.PathValue("phone"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"phone": tr.Phone, "state": tr.State, "orderId": strPtr(tr.OrderID), "messages": toMessageDTOs(tr.Messages)})
}

func (h *handlers) resetConversation(w http.ResponseWriter, r *http.Request) {
	if err := h.s.Assistant.Reset(r.Context(), r.PathValue("phone")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) sampleMedia(w http.ResponseWriter, r *http.Request) {
	media, err := h.s.Assistant.SampleMedia(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]sampleMediaDTO, 0, len(media))
	for _, m := range media {
		out = append(out, sampleMediaDTO(m))
	}
	writeJSON(w, http.StatusOK, map[string]any{"media": out})
}
