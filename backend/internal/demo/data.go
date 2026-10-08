// Package demo is the fictitious dataset of the hackathon MVP: zones,
// pharmacies, products, stock, clients, doctors and sample prescriptions.
// Nothing here comes from a real system.
package demo

import (
	ocrmock "github.com/farmaenlace/farmi/internal/adapters/out/ocr/mock"
	"github.com/farmaenlace/farmi/internal/core/domain"
)

// Zones the client can choose.
func Zones() []domain.Zone {
	return []domain.Zone{
		{ID: "uio-norte", Label: "Quito — zona norte"},
		{ID: "uio-centro", Label: "Quito — zona centro"},
		{ID: "uio-sur", Label: "Quito — zona sur"},
		{ID: "gye-norte", Label: "Guayaquil — zona norte"},
	}
}

// Pharmacies of both chains.
func Pharmacies() []domain.Pharmacy {
	return []domain.Pharmacy{
		{ID: "med-norte", Name: "Medicity Demo Norte", Chain: domain.ChainMedicity, Address: "Av. Ficticia A 123 y Calle Demo", ZoneID: "uio-norte", Location: domain.GeoPoint{Lat: -0.176, Lng: -78.480}},
		{ID: "eco-norte", Name: "Económicas Demo Norte", Chain: domain.ChainEconomicas, Address: "Av. Ficticia D 456", ZoneID: "uio-norte", Location: domain.GeoPoint{Lat: -0.170, Lng: -78.475}},
		{ID: "eco-centro", Name: "Económicas Demo Centro", Chain: domain.ChainEconomicas, Address: "Dirección ficticia B", ZoneID: "uio-centro", Location: domain.GeoPoint{Lat: -0.220, Lng: -78.512}},
		{ID: "med-centro", Name: "Medicity Demo Centro", Chain: domain.ChainMedicity, Address: "Dirección ficticia C", ZoneID: "uio-centro", Location: domain.GeoPoint{Lat: -0.215, Lng: -78.505}},
		{ID: "med-sur", Name: "Medicity Demo Sur", Chain: domain.ChainMedicity, Address: "Av. Ficticia E 789", ZoneID: "uio-sur", Location: domain.GeoPoint{Lat: -0.300, Lng: -78.545}},
		{ID: "eco-gye", Name: "Económicas Demo Guayaquil", Chain: domain.ChainEconomicas, Address: "Av. Ficticia F 1011", ZoneID: "gye-norte", Location: domain.GeoPoint{Lat: -2.150, Lng: -79.900}},
	}
}

// Medicines of the generic catalog.
func Medicines() []domain.Medicine {
	return []domain.Medicine{
		{Key: "paracetamol-500", Name: "Paracetamol", Concentration: "500 mg", Presentation: "tabletas", RequiresPrescription: false},
		{Key: "amoxicilina-500", Name: "Amoxicilina", Concentration: "500 mg", Presentation: "cápsulas", RequiresPrescription: true},
		{Key: "loratadina-10", Name: "Loratadina", Concentration: "10 mg", Presentation: "tabletas", RequiresPrescription: false},
		{Key: "ibuprofeno-400", Name: "Ibuprofeno", Concentration: "400 mg", Presentation: "tabletas", RequiresPrescription: false},
		{Key: "omeprazol-20", Name: "Omeprazol", Concentration: "20 mg", Presentation: "cápsulas", RequiresPrescription: true},
		{Key: "metformina-850", Name: "Metformina", Concentration: "850 mg", Presentation: "tabletas", RequiresPrescription: true},
	}
}

// Products (brands). Prices are per sale unit: a tablet when SellByUnit,
// otherwise a box of UnitsPerPack.
func Products() []domain.Product {
	unit := func(sku, med, brand, conc, pres, label string, cents int64, rx bool) domain.Product {
		return domain.Product{SKU: sku, MedicineKey: med, Brand: brand, Concentration: conc, Presentation: pres,
			UnitsPerPack: 1, SellByUnit: true, RequiresPrescription: rx, UnitPrice: domain.Cents(cents), UnitLabel: label}
	}
	box := func(sku, med, brand, conc, pres string, perPack int, cents int64, rx bool) domain.Product {
		return domain.Product{SKU: sku, MedicineKey: med, Brand: brand, Concentration: conc, Presentation: pres,
			UnitsPerPack: perPack, SellByUnit: false, RequiresPrescription: rx, UnitPrice: domain.Cents(cents), UnitLabel: "cajas"}
	}
	return []domain.Product{
		unit("PAR-ALFA-500", "paracetamol-500", "Marca Alfa", "500 mg", "tabletas", "tabletas", 40, false),
		unit("PAR-GEN-500", "paracetamol-500", "Genérico Demo", "500 mg", "tabletas", "tabletas", 25, false),
		unit("AMX-BETA-500", "amoxicilina-500", "Marca Beta", "500 mg", "cápsulas", "cápsulas", 55, true),
		box("AMX-DELTA-500", "amoxicilina-500", "Marca Delta", "500 mg", "cápsulas", 21, 950, true),
		unit("LOR-GAMMA-10", "loratadina-10", "Marca Gamma", "10 mg", "tabletas", "tabletas", 20, false),
		box("LOR-CLAR-10", "loratadina-10", "Clarityne Demo", "10 mg", "tabletas", 10, 380, false),
		unit("IBU-OMEGA-400", "ibuprofeno-400", "Marca Omega", "400 mg", "tabletas", "tabletas", 30, false),
		unit("OME-SIGMA-20", "omeprazol-20", "Marca Sigma", "20 mg", "cápsulas", "cápsulas", 35, true),
		unit("MET-THETA-850", "metformina-850", "Marca Theta", "850 mg", "tabletas", "tabletas", 15, true),
		box("MET-GLU-850", "metformina-850", "Glucophage Demo", "850 mg", "tabletas", 30, 690, true),
	}
}

// Stock per pharmacy, arranged so receta-001 is complete in zona norte, split
// in zona centro and incomplete (pickup) in zona sur.
func Stock() []domain.StockLevel {
	type row struct {
		pharmacy, sku string
		qty           int
	}
	rows := []row{
		{"med-norte", "PAR-ALFA-500", 100}, {"med-norte", "PAR-GEN-500", 60}, {"med-norte", "AMX-BETA-500", 50},
		{"med-norte", "AMX-DELTA-500", 10}, {"med-norte", "LOR-GAMMA-10", 40}, {"med-norte", "LOR-CLAR-10", 15},
		{"med-norte", "IBU-OMEGA-400", 30}, {"med-norte", "OME-SIGMA-20", 20}, {"med-norte", "MET-THETA-850", 60},
		{"med-norte", "MET-GLU-850", 5},
		{"eco-norte", "PAR-GEN-500", 80}, {"eco-norte", "LOR-GAMMA-10", 25}, {"eco-norte", "IBU-OMEGA-400", 50},
		{"eco-norte", "MET-THETA-850", 100},
		{"eco-centro", "PAR-ALFA-500", 40}, {"eco-centro", "LOR-GAMMA-10", 30}, {"eco-centro", "IBU-OMEGA-400", 20},
		{"med-centro", "PAR-GEN-500", 5}, {"med-centro", "AMX-BETA-500", 30}, {"med-centro", "LOR-CLAR-10", 10},
		{"med-centro", "OME-SIGMA-20", 25}, {"med-centro", "MET-GLU-850", 8},
		{"med-sur", "PAR-ALFA-500", 50}, {"med-sur", "IBU-OMEGA-400", 10}, {"med-sur", "MET-THETA-850", 20},
		{"eco-gye", "IBU-OMEGA-400", 40}, {"eco-gye", "OME-SIGMA-20", 10}, {"eco-gye", "MET-THETA-850", 90},
		{"eco-gye", "PAR-GEN-500", 30},
	}
	out := make([]domain.StockLevel, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.StockLevel{PharmacyID: r.pharmacy, SKU: r.sku, Stock: r.qty})
	}
	return out
}

// Clients already registered.
func Clients() []domain.Client {
	return []domain.Client{
		{ID: "1712345678", Name: "María Pérez", Phone: "+593991111111"},
		{ID: "0912345678", Name: "Juan López", Phone: "+593992222222"},
	}
}

// Doctors of the fictitious registry.
func Doctors() []domain.Doctor {
	return []domain.Doctor{
		{RegistryID: "MSP-10234", Name: "Dra. Ana Torres", Active: true, EnabledToPrescribe: true},
		{RegistryID: "MSP-20987", Name: "Dr. Luis Andrade", Active: true, EnabledToPrescribe: true},
		{RegistryID: "MSP-30111", Name: "Dr. Pedro Salazar", Active: false, EnabledToPrescribe: true},
		{RegistryID: "MSP-40555", Name: "Dra. Carla Mena", Active: true, EnabledToPrescribe: false},
	}
}

// Samples the simulated phone can send and what the OCR returns for each.
func Samples() []ocrmock.Sample {
	item := func(med, conc, pres string, qty int) domain.PrescribedItem {
		return domain.PrescribedItem{Medicine: med, Concentration: conc, Presentation: pres, Quantity: qty, Unit: pres}
	}
	media := func(id, title, desc, expected string) domain.SampleMedia {
		return domain.SampleMedia{ID: id, Title: title, Description: desc, URL: "/recetas/" + id + ".svg", Expected: expected}
	}
	return []ocrmock.Sample{
		{
			Media: media("receta-001", "Receta válida — Dra. Ana Torres", "Paracetamol, Amoxicilina y Loratadina", "valid"),
			Prescription: domain.Prescription{
				Patient: domain.Patient{Name: "María Pérez", IDNumber: "1712345678"}, IssuedAt: "2026-10-07",
				Doctor:       domain.PrescribingDoctor{Name: "Dra. Ana Torres", RegistryID: "MSP-10234"},
				HasSignature: true, HasStamp: true, Confidence: 0.97,
				Items: []domain.PrescribedItem{item("Paracetamol", "500 mg", "tabletas", 20), item("Amoxicilina", "500 mg", "cápsulas", 21), item("Loratadina", "10 mg", "tabletas", 10)},
			},
		},
		{
			Media: media("receta-002", "Receta válida — Dr. Luis Andrade", "Ibuprofeno, Omeprazol y Metformina", "valid"),
			Prescription: domain.Prescription{
				Patient: domain.Patient{Name: "Juan López", IDNumber: "0912345678"}, IssuedAt: "2026-10-06",
				Doctor:       domain.PrescribingDoctor{Name: "Dr. Luis Andrade", RegistryID: "MSP-20987"},
				HasSignature: true, HasStamp: true, Confidence: 0.95,
				Items: []domain.PrescribedItem{item("Ibuprofeno", "400 mg", "tabletas", 12), item("Omeprazol", "20 mg", "cápsulas", 14), item("Metformina", "850 mg", "tabletas", 30)},
			},
		},
		{
			Media: media("receta-003", "Médico inactivo — Dr. Pedro Salazar", "El médico está registrado pero inactivo", "doctor_inactive"),
			Prescription: domain.Prescription{
				Patient: domain.Patient{Name: "María Pérez", IDNumber: "1712345678"}, IssuedAt: "2026-10-05",
				Doctor:       domain.PrescribingDoctor{Name: "Dr. Pedro Salazar", RegistryID: "MSP-30111"},
				HasSignature: true, HasStamp: true, Confidence: 0.93,
				Items: []domain.PrescribedItem{item("Paracetamol", "500 mg", "tabletas", 10)},
			},
		},
		{
			Media: media("receta-004", "Receta incompleta", "Sin firma ni sello del médico", "incomplete"),
			Prescription: domain.Prescription{
				Patient: domain.Patient{Name: "María Pérez", IDNumber: "1712345678"}, IssuedAt: "2026-10-07",
				Doctor:       domain.PrescribingDoctor{Name: "Dra. Ana Torres", RegistryID: "MSP-10234"},
				HasSignature: false, HasStamp: false, Confidence: 0.90,
				Items: []domain.PrescribedItem{item("Loratadina", "10 mg", "tabletas", 10)},
			},
		},
		{
			Media:     media("receta-005", "Imagen ilegible", "El OCR no puede leer la imagen", "illegible"),
			Illegible: true,
		},
		{
			Media: media("receta-006", "Médico no registrado", "Identificador MSP-99999 inexistente", "doctor_unknown"),
			Prescription: domain.Prescription{
				Patient: domain.Patient{Name: "Juan López", IDNumber: "0912345678"}, IssuedAt: "2026-10-07",
				Doctor:       domain.PrescribingDoctor{Name: "Dr. Desconocido", RegistryID: "MSP-99999"},
				HasSignature: true, HasStamp: true, Confidence: 0.91,
				Items: []domain.PrescribedItem{item("Ibuprofeno", "400 mg", "tabletas", 12)},
			},
		},
	}
}

// Courier is the fictitious delivery person.
func Courier() domain.Courier {
	return domain.Courier{Name: "Carlos Repartidor (ficticio)", Phone: "+593990000000"}
}

// Banks shown by the DeUna simulator.
func Banks() []domain.Bank {
	return []domain.Bank{
		{ID: "banco-demo-1", Name: "Banco Demo Pichincha"},
		{ID: "banco-demo-2", Name: "Banco Demo Guayaquil"},
		{ID: "banco-demo-3", Name: "Banco Demo Pacífico"},
		{ID: "banco-demo-4", Name: "Banco Demo Produbanco"},
	}
}
