package rest

import (
	"net/http"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

func (h *handlers) zones(w http.ResponseWriter, r *http.Request) {
	zones, err := h.s.Availability.Zones(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]zoneDTO, 0, len(zones))
	for _, z := range zones {
		out = append(out, zoneDTO(z))
	}
	writeJSON(w, http.StatusOK, map[string]any{"zones": out})
}

func (h *handlers) pharmacies(w http.ResponseWriter, r *http.Request) {
	list, err := h.s.Availability.Pharmacies(r.Context(), r.URL.Query().Get("zone"))
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]pharmacyDTO, 0, len(list))
	for _, p := range list {
		out = append(out, toPharmacyDTO(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{"pharmacies": out})
}

func (h *handlers) availability(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Zone           string `json:"zone"`
		PrescriptionID string `json:"prescriptionId"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.Zone == "" || req.PrescriptionID == "" {
		writeError(w, domain.NewError(domain.CodeValidation, "zone y prescriptionId son obligatorios"))
		return
	}
	rx, err := h.s.OCR.Find(r.Context(), req.PrescriptionID)
	if err != nil {
		writeError(w, err)
		return
	}
	av, err := h.s.Availability.FindOptions(r.Context(), req.Zone, rx)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAvailabilityDTO(av))
}

func (h *handlers) brands(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PharmacyIDs    []string `json:"pharmacyIds"`
		PrescriptionID string   `json:"prescriptionId"`
	}
	if err := decode(r, &req); err != nil {
		writeError(w, err)
		return
	}
	if len(req.PharmacyIDs) == 0 || req.PrescriptionID == "" {
		writeError(w, domain.NewError(domain.CodeValidation, "pharmacyIds y prescriptionId son obligatorios"))
		return
	}
	rx, err := h.s.OCR.Find(r.Context(), req.PrescriptionID)
	if err != nil {
		writeError(w, err)
		return
	}
	brands, err := h.s.Availability.BrandOptionsAt(r.Context(), req.PharmacyIDs, rx)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"medicines": toMedicineBrandsDTOs(brands)})
}
