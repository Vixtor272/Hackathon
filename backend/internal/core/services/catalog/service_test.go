package catalog_test

import (
	"context"
	"testing"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/testkit"
)

func TestZonaNorteSuggestsOnlyTheStoreWithEverything(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	av, err := app.Availability.FindOptions(context.Background(), "uio-norte", app.Receta(t, "receta-001"))
	if err != nil {
		t.Fatal(err)
	}
	// eco-norte + med-norte would also work, but a single pickup wins.
	if len(av.Options) != 1 {
		t.Fatalf("want 1 option, got %d: %+v", len(av.Options), av.Options)
	}
	if av.Options[0].Kind != domain.OptionSingle || av.Options[0].Pharmacies[0].ID != "med-norte" {
		t.Fatalf("option %+v", av.Options[0])
	}
	if len(av.Missing) != 0 || !av.DeliveryAvailable {
		t.Fatalf("missing %+v delivery %v", av.Missing, av.DeliveryAvailable)
	}
}

func TestZonaCentroNeedsTwoStores(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	av, err := app.Availability.FindOptions(context.Background(), "uio-centro", app.Receta(t, "receta-001"))
	if err != nil {
		t.Fatal(err)
	}
	if len(av.Options) != 1 || av.Options[0].Kind != domain.OptionSplit {
		t.Fatalf("options %+v", av.Options)
	}
	where := map[string]string{}
	for _, c := range av.Options[0].Coverage {
		where[c.Medicine] = c.PharmacyID
	}
	if where["Paracetamol 500 mg"] != "eco-centro" || where["Amoxicilina 500 mg"] != "med-centro" || where["Loratadina 10 mg"] != "eco-centro" {
		t.Fatalf("coverage %+v", where)
	}
}

func TestZonaSurMissingForPickupButDeliverable(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	av, err := app.Availability.FindOptions(context.Background(), "uio-sur", app.Receta(t, "receta-001"))
	if err != nil {
		t.Fatal(err)
	}
	if len(av.Options) != 0 || len(av.Missing) != 2 || !av.DeliveryAvailable {
		t.Fatalf("options %d missing %+v delivery %v", len(av.Options), av.Missing, av.DeliveryAvailable)
	}
}

func TestBrandOptionsCheapestFirstAndBoxes(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	rx := app.Receta(t, "receta-001")
	coverage := []domain.Coverage{
		{MedicineKey: "paracetamol-500", PharmacyID: "med-norte", Quantity: 20},
		{MedicineKey: "amoxicilina-500", PharmacyID: "med-norte", Quantity: 21},
	}
	brands, err := app.Availability.BrandOptions(context.Background(), coverage, rx)
	if err != nil {
		t.Fatal(err)
	}
	if brands[0].Brands[0].Product.SKU != "PAR-GEN-500" || brands[0].Brands[0].Subtotal() != 500 {
		t.Fatalf("paracetamol brands %+v", brands[0].Brands)
	}
	if brands[1].Brands[0].Product.SKU != "AMX-DELTA-500" || brands[1].Brands[0].Quantity != 1 || brands[1].Brands[0].Subtotal() != 950 {
		t.Fatalf("amoxicilina box option %+v", brands[1].Brands[0])
	}
	if len(brands[2].Brands) != 0 {
		t.Fatalf("loratadina has no coverage, want no brands")
	}
}

func TestPlanDeliveryConsolidatesInOneStore(t *testing.T) {
	app := testkit.New(10 * time.Minute)
	plan, err := app.Availability.PlanDelivery(context.Background(), app.Receta(t, "receta-001"))
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Feasible() || len(plan.Coverage) != 3 {
		t.Fatalf("plan %+v", plan)
	}
	for _, c := range plan.Coverage {
		if c.PharmacyID != "med-norte" {
			t.Fatalf("want every origin at med-norte, got %+v", plan.Coverage)
		}
	}
}
