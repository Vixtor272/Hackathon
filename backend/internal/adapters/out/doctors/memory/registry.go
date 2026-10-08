// Package memory is the fictitious medical registry.
package memory

import (
	"context"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

// Registry implements ports.DoctorRegistry.
type Registry struct {
	byID map[string]domain.Doctor
}

// NewRegistry indexes the doctors by registry id.
func NewRegistry(doctors []domain.Doctor) *Registry {
	r := &Registry{byID: make(map[string]domain.Doctor, len(doctors))}
	for _, d := range doctors {
		r.byID[d.RegistryID] = d
	}
	return r
}

// FindByRegistryID looks a doctor up.
func (r *Registry) FindByRegistryID(_ context.Context, id string) (domain.Doctor, bool, error) {
	d, ok := r.byID[id]
	return d, ok, nil
}
