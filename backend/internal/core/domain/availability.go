package domain

// OptionKind tells whether a purchase option uses one store or two.
type OptionKind string

const (
	OptionSingle OptionKind = "single"
	OptionSplit  OptionKind = "split"
)

// Coverage assigns one prescribed medicine to the pharmacy that will supply it.
type Coverage struct {
	MedicineKey string
	Medicine    string
	PharmacyID  string
	Quantity    int // prescribed units
}

// MissingItem is a medicine (or part of its quantity) that cannot be supplied.
type MissingItem struct {
	Medicine  string
	Requested int
	Available int
}

// PharmacyOption is a way to fulfil the whole prescription in a zone.
type PharmacyOption struct {
	ID         string
	Kind       OptionKind
	Label      string
	Pharmacies []Pharmacy
	Coverage   []Coverage
	Missing    []MissingItem
}

// BrandOption is one purchasable brand for a prescribed medicine.
type BrandOption struct {
	Medicine        string // "Paracetamol 500 mg"
	Product         Product
	PharmacyID      string
	PharmacyName    string
	PharmacyAddress string
	Quantity        int // sale units needed for the prescribed quantity
	Available       int // sale units available right now
}

// Subtotal is the cost of the quantity needed.
func (b BrandOption) Subtotal() Money { return b.Product.UnitPrice.Times(b.Quantity) }

// MedicineBrands groups the brand options of one prescribed medicine.
type MedicineBrands struct {
	MedicineKey          string
	Medicine             string
	Requested            int
	Unit                 string
	RequiresPrescription bool
	Brands               []BrandOption
}

// Availability is the answer to "where can I buy this prescription in this zone?".
type Availability struct {
	Zone              Zone
	Options           []PharmacyOption
	Missing           []MissingItem
	DeliveryAvailable bool
}

// DeliveryPlan is the company-internal assignment of origin pharmacies for a
// home delivery. The client never sees the pharmacies.
type DeliveryPlan struct {
	Coverage []Coverage
	Missing  []MissingItem
}

// Feasible reports whether every medicine has an origin.
func (p DeliveryPlan) Feasible() bool { return len(p.Missing) == 0 }
