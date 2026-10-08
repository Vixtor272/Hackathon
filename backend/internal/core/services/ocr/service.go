// Package ocr is the use case that turns a prescription image into structured
// JSON through the PrescriptionExtractor port and keeps the result.
package ocr

import (
	"context"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Service implements ports.OCRService.
type Service struct {
	extractor ports.PrescriptionExtractor
	repo      ports.PrescriptionRepository
	ids       ports.IDGenerator
}

// New wires the OCR use case.
func New(extractor ports.PrescriptionExtractor, repo ports.PrescriptionRepository, ids ports.IDGenerator) *Service {
	return &Service{extractor: extractor, repo: repo, ids: ids}
}

// Extract runs the OCR engine and stores the structured prescription.
func (s *Service) Extract(ctx context.Context, mediaID string) (domain.Prescription, error) {
	if mediaID == "" {
		return domain.Prescription{}, domain.NewError(domain.CodeValidation, "Falta el identificador de la imagen")
	}
	p, err := s.extractor.Extract(ctx, mediaID)
	if err != nil {
		return domain.Prescription{}, err
	}
	p.ID = s.ids.New("rx")
	p.MediaID = mediaID
	if err := s.repo.Save(ctx, p); err != nil {
		return domain.Prescription{}, err
	}
	return p, nil
}

// Find returns a previously extracted prescription.
func (s *Service) Find(ctx context.Context, id string) (domain.Prescription, error) {
	return s.repo.FindByID(ctx, id)
}

// SampleMedia lists the images the simulated phone can send.
func (s *Service) SampleMedia(ctx context.Context) ([]domain.SampleMedia, error) {
	return s.extractor.ListMedia(ctx)
}
