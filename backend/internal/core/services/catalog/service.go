// Package catalog is the AI module that consults the company's external
// catalog API and the inventory to answer "where and how can this
// prescription be bought?".
package catalog

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Service implements ports.AvailabilityService.
type Service struct {
	api       ports.CatalogAPI
	inventory ports.InventoryRepository
	ai        ports.ConversationAI
}

// New wires the external API, the inventory and the matcher.
func New(api ports.CatalogAPI, inventory ports.InventoryRepository, ai ports.ConversationAI) *Service {
	return &Service{api: api, inventory: inventory, ai: ai}
}

// requirement is one prescription line resolved against the catalog.
type requirement struct {
	item     domain.PrescribedItem
	medicine domain.Medicine
	products []domain.Product
	matched  bool
}

func (r requirement) label() string {
	if r.matched {
		return r.medicine.Label()
	}
	return r.item.Label()
}

func (r requirement) key() string {
	if r.matched {
		return r.medicine.Key
	}
	return "unknown:" + strings.ToLower(r.item.Label())
}

func (r requirement) requiresRx() bool { return r.matched && r.medicine.RequiresPrescription }

// Zones lists where the client can buy.
func (s *Service) Zones(ctx context.Context) ([]domain.Zone, error) { return s.api.ListZones(ctx) }

// Pharmacies lists the stores of a zone.
func (s *Service) Pharmacies(ctx context.Context, zoneID string) ([]domain.Pharmacy, error) {
	return s.api.ListPharmacies(ctx, zoneID)
}

// FindOptions builds the purchase options of a zone: first the stores that
// cover the whole prescription, then pairs (a main store plus the nearest one
// that has what it lacks), and finally what nobody in the zone can supply.
func (s *Service) FindOptions(ctx context.Context, zoneID string, p domain.Prescription) (domain.Availability, error) {
	zone, err := s.api.FindZone(ctx, zoneID)
	if err != nil {
		return domain.Availability{}, err
	}
	reqs, err := s.resolve(ctx, p)
	if err != nil {
		return domain.Availability{}, err
	}
	pharmacies, err := s.api.ListPharmacies(ctx, zoneID)
	if err != nil {
		return domain.Availability{}, err
	}
	units, err := s.supplyMatrix(ctx, pharmacies, reqs)
	if err != nil {
		return domain.Availability{}, err
	}
	av := domain.Availability{Zone: zone}
	av.Options = append(av.Options, singleStoreOptions(pharmacies, reqs, units)...)
	av.Options = append(av.Options, splitStoreOptions(pharmacies, reqs, units)...)
	for i := range av.Options {
		av.Options[i].ID = fmt.Sprintf("opt_%d", i+1)
	}
	av.Missing = missingIn(pharmacies, reqs, units)
	plan, err := s.PlanDelivery(ctx, p)
	if err != nil {
		return domain.Availability{}, err
	}
	av.DeliveryAvailable = plan.Feasible()
	return av, nil
}

// PlanDelivery assigns origin pharmacies company-wide, preferring the store
// that covers the most medicines so the order is consolidated from few points.
func (s *Service) PlanDelivery(ctx context.Context, p domain.Prescription) (domain.DeliveryPlan, error) {
	reqs, err := s.resolve(ctx, p)
	if err != nil {
		return domain.DeliveryPlan{}, err
	}
	pharmacies, err := s.api.ListPharmacies(ctx, "")
	if err != nil {
		return domain.DeliveryPlan{}, err
	}
	units, err := s.supplyMatrix(ctx, pharmacies, reqs)
	if err != nil {
		return domain.DeliveryPlan{}, err
	}
	remaining := map[int]bool{}
	for i := range reqs {
		remaining[i] = true
	}
	var plan domain.DeliveryPlan
	for len(remaining) > 0 {
		best, bestCount := "", 0
		for _, ph := range pharmacies {
			count := 0
			for i := range remaining {
				if covers(units[ph.ID][i], reqs[i]) {
					count++
				}
			}
			if count > bestCount {
				best, bestCount = ph.ID, count
			}
		}
		if bestCount == 0 {
			break
		}
		for i := 0; i < len(reqs); i++ {
			if remaining[i] && covers(units[best][i], reqs[i]) {
				plan.Coverage = append(plan.Coverage, coverageOf(reqs[i], best))
				delete(remaining, i)
			}
		}
	}
	for i := 0; i < len(reqs); i++ {
		if remaining[i] {
			plan.Missing = append(plan.Missing, missingItem(reqs[i], maxUnits(pharmacies, units, i)))
		}
	}
	return plan, nil
}

// BrandOptions lists, for every medicine, the brands its assigned pharmacy
// can supply in the prescribed quantity, cheapest first.
func (s *Service) BrandOptions(ctx context.Context, coverage []domain.Coverage, p domain.Prescription) ([]domain.MedicineBrands, error) {
	reqs, err := s.resolve(ctx, p)
	if err != nil {
		return nil, err
	}
	byKey := map[string]domain.Coverage{}
	for _, c := range coverage {
		byKey[c.MedicineKey] = c
	}
	out := make([]domain.MedicineBrands, 0, len(reqs))
	for _, r := range reqs {
		mb := domain.MedicineBrands{MedicineKey: r.key(), Medicine: r.label(), Requested: r.item.Quantity,
			Unit: r.item.Unit, RequiresPrescription: r.requiresRx()}
		if cov, ok := byKey[r.key()]; ok {
			brands, err := s.brandsAt(ctx, cov.PharmacyID, r)
			if err != nil {
				return nil, err
			}
			mb.Brands = brands
		}
		out = append(out, mb)
	}
	return out, nil
}

// BrandOptionsAt derives the coverage from a list of stores (first store that
// can supply each medicine wins) and then lists the brands.
func (s *Service) BrandOptionsAt(ctx context.Context, pharmacyIDs []string, p domain.Prescription) ([]domain.MedicineBrands, error) {
	reqs, err := s.resolve(ctx, p)
	if err != nil {
		return nil, err
	}
	var coverage []domain.Coverage
	for _, r := range reqs {
		for _, id := range pharmacyIDs {
			u, err := s.supply(ctx, id, r)
			if err != nil {
				return nil, err
			}
			if covers(u, r) {
				coverage = append(coverage, coverageOf(r, id))
				break
			}
		}
	}
	return s.BrandOptions(ctx, coverage, p)
}

func (s *Service) brandsAt(ctx context.Context, pharmacyID string, r requirement) ([]domain.BrandOption, error) {
	ph, err := s.api.FindPharmacy(ctx, pharmacyID)
	if err != nil {
		return nil, err
	}
	var brands []domain.BrandOption
	for _, prod := range r.products {
		need := prod.SaleQuantityFor(r.item.Quantity)
		avail, err := s.inventory.Available(ctx, pharmacyID, prod.SKU)
		if err != nil {
			return nil, err
		}
		if avail < need {
			continue
		}
		brands = append(brands, domain.BrandOption{Medicine: r.label(), Product: prod, PharmacyID: ph.ID,
			PharmacyName: ph.Name, PharmacyAddress: ph.Address, Quantity: need, Available: avail})
	}
	sort.SliceStable(brands, func(i, j int) bool {
		if brands[i].Subtotal() != brands[j].Subtotal() {
			return brands[i].Subtotal() < brands[j].Subtotal()
		}
		return brands[i].Product.Brand < brands[j].Product.Brand
	})
	return brands, nil
}

// resolve maps every prescription line onto a catalog medicine and its brands.
func (s *Service) resolve(ctx context.Context, p domain.Prescription) ([]requirement, error) {
	medicines, err := s.api.ListMedicines(ctx)
	if err != nil {
		return nil, err
	}
	reqs := make([]requirement, 0, len(p.Items))
	for _, it := range p.Items {
		r := requirement{item: it}
		if med, ok := s.ai.MatchMedicine(ctx, it, medicines); ok {
			r.medicine, r.matched = med, true
			if r.products, err = s.api.ProductsByMedicine(ctx, med.Key); err != nil {
				return nil, err
			}
		}
		reqs = append(reqs, r)
	}
	return reqs, nil
}

// supply is the number of prescribed units one store can sell of a medicine
// using its best single brand (the client picks one brand per medicine).
func (s *Service) supply(ctx context.Context, pharmacyID string, r requirement) (int, error) {
	best := 0
	for _, prod := range r.products {
		avail, err := s.inventory.Available(ctx, pharmacyID, prod.SKU)
		if err != nil {
			return 0, err
		}
		if u := prod.UnitsIn(avail); u > best {
			best = u
		}
	}
	return best, nil
}

// supplyMatrix is supply() for every store × requirement.
func (s *Service) supplyMatrix(ctx context.Context, pharmacies []domain.Pharmacy, reqs []requirement) (map[string][]int, error) {
	m := make(map[string][]int, len(pharmacies))
	for _, ph := range pharmacies {
		row := make([]int, len(reqs))
		for i, r := range reqs {
			u, err := s.supply(ctx, ph.ID, r)
			if err != nil {
				return nil, err
			}
			row[i] = u
		}
		m[ph.ID] = row
	}
	return m, nil
}

func covers(units int, r requirement) bool { return r.matched && units >= r.item.Quantity }

func coverageOf(r requirement, pharmacyID string) domain.Coverage {
	return domain.Coverage{MedicineKey: r.key(), Medicine: r.label(), PharmacyID: pharmacyID, Quantity: r.item.Quantity}
}

func missingItem(r requirement, available int) domain.MissingItem {
	return domain.MissingItem{Medicine: r.label(), Requested: r.item.Quantity, Available: available}
}

func maxUnits(pharmacies []domain.Pharmacy, units map[string][]int, i int) int {
	best := 0
	for _, ph := range pharmacies {
		if u := units[ph.ID][i]; u > best {
			best = u
		}
	}
	return best
}

func coveredIdx(ph domain.Pharmacy, reqs []requirement, units map[string][]int) (covered, uncovered []int) {
	for i, r := range reqs {
		if covers(units[ph.ID][i], r) {
			covered = append(covered, i)
		} else {
			uncovered = append(uncovered, i)
		}
	}
	return covered, uncovered
}

func singleStoreOptions(pharmacies []domain.Pharmacy, reqs []requirement, units map[string][]int) []domain.PharmacyOption {
	var out []domain.PharmacyOption
	for _, ph := range pharmacies {
		covered, uncovered := coveredIdx(ph, reqs, units)
		if len(uncovered) > 0 || len(covered) == 0 {
			continue
		}
		opt := domain.PharmacyOption{Kind: domain.OptionSingle, Label: ph.Name + " — toda la receta", Pharmacies: []domain.Pharmacy{ph}}
		for _, i := range covered {
			opt.Coverage = append(opt.Coverage, coverageOf(reqs[i], ph.ID))
		}
		out = append(out, opt)
	}
	return out
}

func splitStoreOptions(pharmacies []domain.Pharmacy, reqs []requirement, units map[string][]int) []domain.PharmacyOption {
	primaries := append([]domain.Pharmacy(nil), pharmacies...)
	sort.SliceStable(primaries, func(i, j int) bool {
		ci, _ := coveredIdx(primaries[i], reqs, units)
		cj, _ := coveredIdx(primaries[j], reqs, units)
		if len(ci) != len(cj) {
			return len(ci) > len(cj)
		}
		return primaries[i].Name < primaries[j].Name
	})
	seen := map[string]bool{}
	var out []domain.PharmacyOption
	for _, primary := range primaries {
		covered, uncovered := coveredIdx(primary, reqs, units)
		if len(covered) == 0 || len(uncovered) == 0 {
			continue
		}
		var candidates []domain.Pharmacy
		for _, other := range pharmacies {
			if other.ID == primary.ID {
				continue
			}
			ok := true
			for _, i := range uncovered {
				ok = ok && covers(units[other.ID][i], reqs[i])
			}
			if ok {
				candidates = append(candidates, other)
			}
		}
		if len(candidates) == 0 {
			continue
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			return primary.DistanceKm(candidates[i]) < primary.DistanceKm(candidates[j])
		})
		secondary := candidates[0]
		key := pairKey(primary.ID, secondary.ID)
		if seen[key] {
			continue
		}
		seen[key] = true
		opt := domain.PharmacyOption{Kind: domain.OptionSplit,
			Label:      primary.Name + " + " + secondary.Name + " — receta completa entre ambos locales",
			Pharmacies: []domain.Pharmacy{primary, secondary}}
		for _, i := range covered {
			opt.Coverage = append(opt.Coverage, coverageOf(reqs[i], primary.ID))
		}
		for _, i := range uncovered {
			opt.Coverage = append(opt.Coverage, coverageOf(reqs[i], secondary.ID))
		}
		out = append(out, opt)
	}
	return out
}

func pairKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "|" + b
}

func missingIn(pharmacies []domain.Pharmacy, reqs []requirement, units map[string][]int) []domain.MissingItem {
	var out []domain.MissingItem
	for i, r := range reqs {
		best := maxUnits(pharmacies, units, i)
		if !covers(best, r) {
			out = append(out, missingItem(r, best))
		}
	}
	return out
}
