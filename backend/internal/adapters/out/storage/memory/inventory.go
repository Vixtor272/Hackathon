package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

type stockKey struct{ pharmacyID, sku string }

// InventoryRepository implements the demo stock model:
// availability = stock − reservations; reserve/release move reservations,
// commit consumes both.
type InventoryRepository struct {
	mu     sync.Mutex
	levels map[stockKey]*domain.StockLevel
}

// NewInventoryRepository seeds the stock of every pharmacy.
func NewInventoryRepository(seed []domain.StockLevel) *InventoryRepository {
	r := &InventoryRepository{levels: map[stockKey]*domain.StockLevel{}}
	for _, s := range seed {
		lvl := s
		r.levels[stockKey{s.PharmacyID, s.SKU}] = &lvl
	}
	return r
}

func (r *InventoryRepository) level(pharmacyID, sku string) *domain.StockLevel {
	if lvl, ok := r.levels[stockKey{pharmacyID, sku}]; ok {
		return lvl
	}
	lvl := &domain.StockLevel{PharmacyID: pharmacyID, SKU: sku}
	r.levels[stockKey{pharmacyID, sku}] = lvl
	return lvl
}

// Available is stock minus reservations.
func (r *InventoryRepository) Available(_ context.Context, pharmacyID, sku string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.level(pharmacyID, sku).Available(), nil
}

// Reserve holds units; fails when availability is short.
func (r *InventoryRepository) Reserve(_ context.Context, pharmacyID, sku string, qty int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	lvl := r.level(pharmacyID, sku)
	if lvl.Available() < qty {
		return domain.NewError(domain.CodeInsufficientStock, "Disponibles %d de %s en %s, solicitadas %d", lvl.Available(), sku, pharmacyID, qty)
	}
	lvl.Reserved += qty
	return nil
}

// Release frees held units.
func (r *InventoryRepository) Release(_ context.Context, pharmacyID, sku string, qty int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	lvl := r.level(pharmacyID, sku)
	lvl.Reserved -= qty
	if lvl.Reserved < 0 {
		lvl.Reserved = 0
	}
	return nil
}

// Commit turns a reservation into a sale: stock and reservation both drop.
func (r *InventoryRepository) Commit(_ context.Context, pharmacyID, sku string, qty int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	lvl := r.level(pharmacyID, sku)
	lvl.Stock -= qty
	lvl.Reserved -= qty
	if lvl.Reserved < 0 {
		lvl.Reserved = 0
	}
	if lvl.Stock < 0 {
		lvl.Stock = 0
	}
	return nil
}

// Levels snapshots the whole inventory.
func (r *InventoryRepository) Levels(_ context.Context) ([]domain.StockLevel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.StockLevel, 0, len(r.levels))
	for _, lvl := range r.levels {
		out = append(out, *lvl)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].PharmacyID != out[j].PharmacyID {
			return out[i].PharmacyID < out[j].PharmacyID
		}
		return out[i].SKU < out[j].SKU
	})
	return out, nil
}
