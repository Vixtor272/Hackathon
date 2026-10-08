// Package client exposes registered buyers.
package client

import (
	"context"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Service implements ports.ClientService.
type Service struct {
	repo ports.ClientRepository
}

// New wires the repository.
func New(repo ports.ClientRepository) *Service { return &Service{repo: repo} }

// Find returns a client by cédula.
func (s *Service) Find(ctx context.Context, id string) (domain.Client, error) {
	c, ok, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return domain.Client{}, err
	}
	if !ok {
		return domain.Client{}, domain.NotFound("Cliente", id)
	}
	return c, nil
}
