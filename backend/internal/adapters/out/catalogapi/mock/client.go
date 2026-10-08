// Package mock stands in for the company's external catalog API. It answers
// the same questions a REST client would, from an in-memory dataset.
package mock

import (
	"context"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

// Data is the dataset the fake API serves.
type Data struct {
	Zones      []domain.Zone
	Pharmacies []domain.Pharmacy
	Medicines  []domain.Medicine
	Products   []domain.Product
}

// Client implements ports.CatalogAPI.
type Client struct {
	data       Data
	zones      map[string]domain.Zone
	pharmacies map[string]domain.Pharmacy
	products   map[string]domain.Product
	byMedicine map[string][]domain.Product
}

// NewClient indexes the dataset.
func NewClient(d Data) *Client {
	c := &Client{data: d, zones: map[string]domain.Zone{}, pharmacies: map[string]domain.Pharmacy{},
		products: map[string]domain.Product{}, byMedicine: map[string][]domain.Product{}}
	for _, z := range d.Zones {
		c.zones[z.ID] = z
	}
	for _, p := range d.Pharmacies {
		c.pharmacies[p.ID] = p
	}
	for _, p := range d.Products {
		c.products[p.SKU] = p
		c.byMedicine[p.MedicineKey] = append(c.byMedicine[p.MedicineKey], p)
	}
	return c
}

// ListZones returns every zone.
func (c *Client) ListZones(_ context.Context) ([]domain.Zone, error) {
	return append([]domain.Zone(nil), c.data.Zones...), nil
}

// FindZone looks a zone up.
func (c *Client) FindZone(_ context.Context, id string) (domain.Zone, error) {
	z, ok := c.zones[id]
	if !ok {
		return domain.Zone{}, domain.NotFound("Zona", id)
	}
	return z, nil
}

// ListPharmacies filters by zone ("" = all).
func (c *Client) ListPharmacies(_ context.Context, zoneID string) ([]domain.Pharmacy, error) {
	var out []domain.Pharmacy
	for _, p := range c.data.Pharmacies {
		if zoneID == "" || p.ZoneID == zoneID {
			out = append(out, p)
		}
	}
	return out, nil
}

// FindPharmacy looks a store up.
func (c *Client) FindPharmacy(_ context.Context, id string) (domain.Pharmacy, error) {
	p, ok := c.pharmacies[id]
	if !ok {
		return domain.Pharmacy{}, domain.NotFound("Farmacia", id)
	}
	return p, nil
}

// ListMedicines returns the generic catalog.
func (c *Client) ListMedicines(_ context.Context) ([]domain.Medicine, error) {
	return append([]domain.Medicine(nil), c.data.Medicines...), nil
}

// ProductsByMedicine returns the brands of a medicine.
func (c *Client) ProductsByMedicine(_ context.Context, key string) ([]domain.Product, error) {
	return append([]domain.Product(nil), c.byMedicine[key]...), nil
}

// FindProduct looks a SKU up.
func (c *Client) FindProduct(_ context.Context, sku string) (domain.Product, error) {
	p, ok := c.products[sku]
	if !ok {
		return domain.Product{}, domain.NotFound("Producto", sku)
	}
	return p, nil
}
