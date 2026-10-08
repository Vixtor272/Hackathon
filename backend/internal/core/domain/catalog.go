package domain

import (
	"math"
	"strings"
)

// Zone is a city area the client can buy in.
type Zone struct {
	ID    string
	Label string
}

// Chain is one of the pharmacy brands of the company.
type Chain string

const (
	ChainMedicity   Chain = "Medicity"
	ChainEconomicas Chain = "Farmacias Económicas"
)

// GeoPoint is a WGS84 coordinate used to pick the nearest pharmacy.
type GeoPoint struct {
	Lat float64
	Lng float64
}

// Pharmacy is a physical store.
type Pharmacy struct {
	ID       string
	Name     string
	Chain    Chain
	Address  string
	ZoneID   string
	Location GeoPoint
}

// DistanceKm returns the great-circle distance between two pharmacies.
func (p Pharmacy) DistanceKm(o Pharmacy) float64 {
	const earthRadiusKm = 6371.0
	toRad := func(deg float64) float64 { return deg * math.Pi / 180 }
	dLat := toRad(o.Location.Lat - p.Location.Lat)
	dLng := toRad(o.Location.Lng - p.Location.Lng)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(p.Location.Lat))*math.Cos(toRad(o.Location.Lat))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthRadiusKm * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// Medicine is a generic drug + concentration the catalog knows about.
type Medicine struct {
	Key                  string
	Name                 string
	Concentration        string
	Presentation         string
	RequiresPrescription bool
}

// Label renders "Paracetamol 500 mg".
func (m Medicine) Label() string {
	return strings.TrimSpace(m.Name + " " + m.Concentration)
}

// Product is a sellable brand of a medicine. UnitPrice is the price of one
// sale unit: a tablet when SellByUnit, otherwise a box of UnitsPerPack.
type Product struct {
	SKU                  string
	MedicineKey          string
	Brand                string
	Concentration        string
	Presentation         string
	UnitsPerPack         int
	SellByUnit           bool
	RequiresPrescription bool
	UnitPrice            Money
	UnitLabel            string
}

// SaleQuantityFor converts prescribed units into sale units (boxes round up).
func (p Product) SaleQuantityFor(units int) int {
	if p.SellByUnit || p.UnitsPerPack <= 1 {
		return units
	}
	return (units + p.UnitsPerPack - 1) / p.UnitsPerPack
}

// UnitsIn converts sale units back into tablets/capsules.
func (p Product) UnitsIn(saleQty int) int {
	if p.SellByUnit || p.UnitsPerPack <= 1 {
		return saleQty
	}
	return saleQty * p.UnitsPerPack
}

// StockLevel is the inventory of one SKU in one pharmacy.
// Demo model: availability = stock − reservations.
type StockLevel struct {
	PharmacyID string
	SKU        string
	Stock      int
	Reserved   int
}

// Available is what can still be sold or reserved.
func (s StockLevel) Available() int { return s.Stock - s.Reserved }
