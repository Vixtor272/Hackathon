// Package prescription validates an OCR result: required fields plus the
// doctor's standing in the (fictitious) registry.
package prescription

import (
	"context"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Validator implements ports.PrescriptionValidator.
type Validator struct {
	doctors ports.DoctorRegistry
}

// NewValidator wires the registry.
func NewValidator(doctors ports.DoctorRegistry) *Validator {
	return &Validator{doctors: doctors}
}

// Validate never returns a domain error for a bad prescription; the verdict
// travels in the result so the caller can show every reason at once.
func (v *Validator) Validate(ctx context.Context, p domain.Prescription) (domain.ValidationResult, error) {
	res := domain.ValidationResult{Errors: p.MissingFields()}
	res.Doctor = domain.DoctorCheck{RegistryID: p.Doctor.RegistryID, Name: p.Doctor.Name}
	if p.Doctor.RegistryID != "" {
		doc, found, err := v.doctors.FindByRegistryID(ctx, p.Doctor.RegistryID)
		if err != nil {
			return domain.ValidationResult{}, err
		}
		res.Doctor.Registered = found
		switch {
		case !found:
			res.Errors = append(res.Errors, "El médico "+p.Doctor.RegistryID+" no está registrado")
		default:
			res.Doctor.Name = doc.Name
			res.Doctor.Active = doc.Active
			res.Doctor.EnabledToPrescribe = doc.EnabledToPrescribe
			if !doc.Active {
				res.Errors = append(res.Errors, "El médico "+doc.Name+" no está activo en el registro")
			}
			if !doc.EnabledToPrescribe {
				res.Errors = append(res.Errors, "El médico "+doc.Name+" no está habilitado para prescribir en Ecuador")
			}
		}
	}
	res.Valid = len(res.Errors) == 0
	return res, nil
}
