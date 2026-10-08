// Package rest is the driving HTTP adapter: it maps the REST contract in
// docs/API.md onto the driving ports and back.
package rest

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError maps business codes to HTTP statuses; anything else is a 500.
func writeError(w http.ResponseWriter, err error) {
	var de *domain.Error
	if errors.As(err, &de) {
		writeJSON(w, statusFor(de.Code), errorEnvelope{errorBody{string(de.Code), de.Message}})
		return
	}
	log.Printf("internal error: %v", err)
	writeJSON(w, http.StatusInternalServerError, errorEnvelope{errorBody{"INTERNAL", "Error interno del servidor"}})
}

func statusFor(code domain.ErrorCode) int {
	switch code {
	case domain.CodeValidation:
		return http.StatusBadRequest
	case domain.CodeNotFound:
		return http.StatusNotFound
	case domain.CodeInvalidState, domain.CodeInsufficientStock, domain.CodeRxIncreaseNotAllowed,
		domain.CodeReservationExpired, domain.CodeEmptyCart, domain.CodePaymentInvalidated:
		return http.StatusConflict
	case domain.CodeOCRIllegible, domain.CodePrescriptionInvalid:
		return http.StatusUnprocessableEntity
	}
	return http.StatusInternalServerError
}

func decode(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return domain.NewError(domain.CodeValidation, "Cuerpo JSON inválido: %v", err)
	}
	return nil
}
