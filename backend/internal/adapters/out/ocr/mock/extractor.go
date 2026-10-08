// Package mock is the simulated OCR engine: each sample image id maps to a
// preloaded structured prescription (or to an "illegible" outcome).
package mock

import (
	"context"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

// Sample pairs a media id with the JSON the OCR would return for it.
type Sample struct {
	Media        domain.SampleMedia
	Prescription domain.Prescription
	Illegible    bool
}

// Extractor implements ports.PrescriptionExtractor.
type Extractor struct {
	byID  map[string]Sample
	media []domain.SampleMedia
}

// NewExtractor indexes the samples.
func NewExtractor(samples []Sample) *Extractor {
	e := &Extractor{byID: make(map[string]Sample, len(samples))}
	for _, s := range samples {
		e.byID[s.Media.ID] = s
		e.media = append(e.media, s.Media)
	}
	return e
}

// Extract returns the preloaded prescription for a sample image.
func (e *Extractor) Extract(_ context.Context, mediaID string) (domain.Prescription, error) {
	s, ok := e.byID[mediaID]
	if !ok {
		return domain.Prescription{}, domain.NotFound("Imagen", mediaID)
	}
	if s.Illegible {
		return domain.Prescription{}, domain.NewError(domain.CodeOCRIllegible, "La imagen %s no pudo ser leída", mediaID)
	}
	p := s.Prescription
	p.Items = append([]domain.PrescribedItem(nil), s.Prescription.Items...)
	return p, nil
}

// ListMedia lists the sample images.
func (e *Extractor) ListMedia(_ context.Context) ([]domain.SampleMedia, error) {
	return append([]domain.SampleMedia(nil), e.media...), nil
}
