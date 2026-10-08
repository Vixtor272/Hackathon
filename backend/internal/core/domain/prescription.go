package domain

import (
	"fmt"
	"strings"
)

// Patient as read from the prescription.
type Patient struct {
	Name     string
	IDNumber string
}

// PrescribingDoctor as read from the prescription.
type PrescribingDoctor struct {
	Name       string
	RegistryID string
}

// PrescribedItem is one medicine line of the prescription.
type PrescribedItem struct {
	Medicine      string
	Concentration string
	Presentation  string
	Quantity      int
	Unit          string
}

// Label joins medicine and concentration: "Paracetamol 500 mg".
func (i PrescribedItem) Label() string {
	return strings.TrimSpace(i.Medicine + " " + i.Concentration)
}

// Prescription is the structured output of the OCR module.
type Prescription struct {
	ID           string
	MediaID      string
	Patient      Patient
	IssuedAt     string
	Doctor       PrescribingDoctor
	HasSignature bool
	HasStamp     bool
	Confidence   float64
	Items        []PrescribedItem
}

// MissingFields lists the required data the demo expects on every prescription.
func (p Prescription) MissingFields() []string {
	var missing []string
	if strings.TrimSpace(p.Patient.Name) == "" {
		missing = append(missing, "Falta el nombre del paciente")
	}
	if strings.TrimSpace(p.IssuedAt) == "" {
		missing = append(missing, "Falta la fecha de emisión")
	}
	if strings.TrimSpace(p.Doctor.Name) == "" {
		missing = append(missing, "Falta el nombre del médico")
	}
	if strings.TrimSpace(p.Doctor.RegistryID) == "" {
		missing = append(missing, "Falta el identificador del médico")
	}
	if !p.HasSignature {
		missing = append(missing, "Falta la firma del médico")
	}
	if !p.HasStamp {
		missing = append(missing, "Falta el sello del médico")
	}
	if len(p.Items) == 0 {
		missing = append(missing, "No se detectaron medicamentos")
	}
	for _, it := range p.Items {
		if it.Quantity <= 0 {
			missing = append(missing, fmt.Sprintf("Cantidad inválida para %s", it.Label()))
		}
	}
	return missing
}

// DoctorCheck is the outcome of looking the prescribing doctor up in the registry.
type DoctorCheck struct {
	RegistryID         string
	Name               string
	Registered         bool
	Active             bool
	EnabledToPrescribe bool
}

// ValidationResult tells whether a prescription can be used to buy.
type ValidationResult struct {
	Valid  bool
	Errors []string
	Doctor DoctorCheck
}

// SampleMedia describes a prescription image the simulated phone can send.
type SampleMedia struct {
	ID          string
	Title       string
	Description string
	URL         string
	Expected    string
}
