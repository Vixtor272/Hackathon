package rest

import "net/http"

func (h *handlers) extract(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MediaID string `json:"mediaId"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	p, err := h.s.OCR.Extract(r.Context(), req.MediaID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"prescription": toPrescriptionDTO(p)})
}

func (h *handlers) validate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prescription prescriptionDTO `json:"prescription"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	res, err := h.s.Validator.Validate(r.Context(), req.Prescription.toDomain())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toValidationDTO(res))
}
