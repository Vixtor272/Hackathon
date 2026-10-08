package rest

import (
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

func ts(t time.Time) string { return t.UTC().Format(time.RFC3339) }

func tsPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := ts(*t)
	return &s
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// --- WhatsApp ---------------------------------------------------------------

type messageDTO struct {
	ID        string             `json:"id"`
	Direction string             `json:"direction"`
	Type      string             `json:"type"`
	Text      string             `json:"text"`
	MediaID   string             `json:"mediaId,omitempty"`
	MediaURL  string             `json:"mediaUrl,omitempty"`
	Link      string             `json:"link,omitempty"`
	Options   []messageOptionDTO `json:"options,omitempty"`
	At        string             `json:"at"`
}

type messageOptionDTO struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func toMessageDTO(m domain.Message) messageDTO {
	d := messageDTO{ID: m.ID, Direction: string(m.Direction), Type: string(m.Type), Text: m.Text,
		MediaID: m.MediaID, MediaURL: m.MediaURL, Link: m.Link, At: ts(m.At)}
	for _, o := range m.Options {
		d.Options = append(d.Options, messageOptionDTO(o))
	}
	return d
}

func toMessageDTOs(ms []domain.Message) []messageDTO {
	out := make([]messageDTO, 0, len(ms))
	for _, m := range ms {
		out = append(out, toMessageDTO(m))
	}
	return out
}

type sampleMediaDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Expected    string `json:"expected"`
}

// --- OCR / prescriptions ----------------------------------------------------

type patientDTO struct {
	Name     string `json:"name"`
	IDNumber string `json:"idNumber"`
}

type doctorRefDTO struct {
	Name       string `json:"name"`
	RegistryID string `json:"registryId"`
}

type prescribedItemDTO struct {
	Medicine      string `json:"medicine"`
	Concentration string `json:"concentration"`
	Presentation  string `json:"presentation"`
	Quantity      int    `json:"quantity"`
	Unit          string `json:"unit"`
}

type prescriptionDTO struct {
	ID           string              `json:"id"`
	MediaID      string              `json:"mediaId"`
	Patient      patientDTO          `json:"patient"`
	IssuedAt     string              `json:"issuedAt"`
	Doctor       doctorRefDTO        `json:"doctor"`
	HasSignature bool                `json:"hasSignature"`
	HasStamp     bool                `json:"hasStamp"`
	Confidence   float64             `json:"confidence"`
	Items        []prescribedItemDTO `json:"items"`
}

func toPrescriptionDTO(p domain.Prescription) prescriptionDTO {
	d := prescriptionDTO{ID: p.ID, MediaID: p.MediaID, Patient: patientDTO(p.Patient), IssuedAt: p.IssuedAt,
		Doctor: doctorRefDTO(p.Doctor), HasSignature: p.HasSignature, HasStamp: p.HasStamp, Confidence: p.Confidence,
		Items: make([]prescribedItemDTO, 0, len(p.Items))}
	for _, it := range p.Items {
		d.Items = append(d.Items, prescribedItemDTO(it))
	}
	return d
}

func (d prescriptionDTO) toDomain() domain.Prescription {
	p := domain.Prescription{ID: d.ID, MediaID: d.MediaID, Patient: domain.Patient(d.Patient), IssuedAt: d.IssuedAt,
		Doctor: domain.PrescribingDoctor(d.Doctor), HasSignature: d.HasSignature, HasStamp: d.HasStamp, Confidence: d.Confidence}
	for _, it := range d.Items {
		p.Items = append(p.Items, domain.PrescribedItem(it))
	}
	return p
}

type doctorCheckDTO struct {
	RegistryID         string `json:"registryId"`
	Name               string `json:"name"`
	Registered         bool   `json:"registered"`
	Active             bool   `json:"active"`
	EnabledToPrescribe bool   `json:"enabledToPrescribe"`
}

type validationDTO struct {
	Valid  bool           `json:"valid"`
	Errors []string       `json:"errors"`
	Doctor doctorCheckDTO `json:"doctor"`
}

func toValidationDTO(v domain.ValidationResult) validationDTO {
	errs := v.Errors
	if errs == nil {
		errs = []string{}
	}
	return validationDTO{Valid: v.Valid, Errors: errs, Doctor: doctorCheckDTO(v.Doctor)}
}

// --- Catalog ----------------------------------------------------------------

type zoneDTO struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type pharmacyDTO struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Chain   string `json:"chain"`
	Address string `json:"address"`
	Zone    string `json:"zone"`
}

func toPharmacyDTO(p domain.Pharmacy) pharmacyDTO {
	return pharmacyDTO{ID: p.ID, Name: p.Name, Chain: string(p.Chain), Address: p.Address, Zone: p.ZoneID}
}

type coverageDTO struct {
	Medicine   string `json:"medicine"`
	PharmacyID string `json:"pharmacyId"`
	Quantity   int    `json:"quantity"`
}

type missingDTO struct {
	Medicine  string `json:"medicine"`
	Requested int    `json:"requested"`
	Available int    `json:"available"`
}

func toMissingDTOs(ms []domain.MissingItem) []missingDTO {
	out := make([]missingDTO, 0, len(ms))
	for _, m := range ms {
		out = append(out, missingDTO(m))
	}
	return out
}

type optionDTO struct {
	ID         string        `json:"id"`
	Kind       string        `json:"kind"`
	Label      string        `json:"label"`
	Pharmacies []pharmacyDTO `json:"pharmacies"`
	Coverage   []coverageDTO `json:"coverage"`
	Missing    []missingDTO  `json:"missing"`
}

type availabilityDTO struct {
	Options           []optionDTO  `json:"options"`
	Missing           []missingDTO `json:"missing"`
	DeliveryAvailable bool         `json:"deliveryAvailable"`
}

func toAvailabilityDTO(a domain.Availability) availabilityDTO {
	d := availabilityDTO{Options: make([]optionDTO, 0, len(a.Options)), Missing: toMissingDTOs(a.Missing), DeliveryAvailable: a.DeliveryAvailable}
	for _, o := range a.Options {
		od := optionDTO{ID: o.ID, Kind: string(o.Kind), Label: o.Label, Pharmacies: []pharmacyDTO{}, Coverage: []coverageDTO{}, Missing: toMissingDTOs(o.Missing)}
		for _, p := range o.Pharmacies {
			od.Pharmacies = append(od.Pharmacies, toPharmacyDTO(p))
		}
		for _, c := range o.Coverage {
			od.Coverage = append(od.Coverage, coverageDTO{Medicine: c.Medicine, PharmacyID: c.PharmacyID, Quantity: c.Quantity})
		}
		d.Options = append(d.Options, od)
	}
	return d
}

type brandDTO struct {
	SKU                  string  `json:"sku"`
	Brand                string  `json:"brand"`
	Concentration        string  `json:"concentration"`
	Presentation         string  `json:"presentation"`
	SellByUnit           bool    `json:"sellByUnit"`
	UnitsPerPack         int     `json:"unitsPerPack"`
	UnitLabel            string  `json:"unitLabel"`
	RequiresPrescription bool    `json:"requiresPrescription"`
	UnitPrice            float64 `json:"unitPrice"`
	Quantity             int     `json:"quantity"`
	Subtotal             float64 `json:"subtotal"`
	PharmacyID           string  `json:"pharmacyId"`
	Available            int     `json:"available"`
}

type medicineBrandsDTO struct {
	Medicine             string     `json:"medicine"`
	Requested            int        `json:"requested"`
	RequiresPrescription bool       `json:"requiresPrescription"`
	Brands               []brandDTO `json:"brands"`
}

func toMedicineBrandsDTOs(ms []domain.MedicineBrands) []medicineBrandsDTO {
	out := make([]medicineBrandsDTO, 0, len(ms))
	for _, m := range ms {
		d := medicineBrandsDTO{Medicine: m.Medicine, Requested: m.Requested, RequiresPrescription: m.RequiresPrescription, Brands: []brandDTO{}}
		for _, b := range m.Brands {
			d.Brands = append(d.Brands, brandDTO{SKU: b.Product.SKU, Brand: b.Product.Brand, Concentration: b.Product.Concentration,
				Presentation: b.Product.Presentation, SellByUnit: b.Product.SellByUnit, UnitsPerPack: b.Product.UnitsPerPack,
				UnitLabel: b.Product.UnitLabel, RequiresPrescription: b.Product.RequiresPrescription,
				UnitPrice: b.Product.UnitPrice.Float(), Quantity: b.Quantity, Subtotal: b.Subtotal().Float(),
				PharmacyID: b.PharmacyID, Available: b.Available})
		}
		out = append(out, d)
	}
	return out
}

// --- Orders -----------------------------------------------------------------

type orderItemDTO struct {
	ID                   string  `json:"id"`
	SKU                  string  `json:"sku"`
	Medicine             string  `json:"medicine"`
	Brand                string  `json:"brand"`
	Presentation         string  `json:"presentation"`
	UnitLabel            string  `json:"unitLabel"`
	PharmacyID           string  `json:"pharmacyId"`
	PharmacyName         string  `json:"pharmacyName"`
	Quantity             int     `json:"quantity"`
	PrescribedQuantity   int     `json:"prescribedQuantity"`
	UnitPrice            float64 `json:"unitPrice"`
	Subtotal             float64 `json:"subtotal"`
	RequiresPrescription bool    `json:"requiresPrescription"`
	CanIncrease          bool    `json:"canIncrease"`
	CanDecrease          bool    `json:"canDecrease"`
}

type fulfillmentItemDTO struct {
	ID       string `json:"id"`
	Medicine string `json:"medicine"`
	Brand    string `json:"brand"`
	Quantity int    `json:"quantity"`
}

type fulfillmentDTO struct {
	PharmacyID   string               `json:"pharmacyId"`
	PharmacyName string               `json:"pharmacyName"`
	Address      string               `json:"address"`
	Status       string               `json:"status"`
	ETA          *string              `json:"eta"`
	ETAMinutes   int                  `json:"etaMinutes"`
	Items        []fulfillmentItemDTO `json:"items"`
}

type courierDTO struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type deliveryDTO struct {
	Status     string     `json:"status"`
	ETA        *string    `json:"eta"`
	ETAMinutes int        `json:"etaMinutes"`
	Courier    courierDTO `json:"courier"`
}

type paymentRefDTO struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Status string `json:"status"`
}

type reservationDTO struct {
	ExpiresAt   string `json:"expiresAt"`
	SecondsLeft int    `json:"secondsLeft"`
	Active      bool   `json:"active"`
}

type orderDTO struct {
	ID              string           `json:"id"`
	Code            string           `json:"code"`
	Status          string           `json:"status"`
	ClientID        string           `json:"clientId"`
	ClientName      string           `json:"clientName"`
	Phone           string           `json:"phone"`
	PrescriptionID  string           `json:"prescriptionId"`
	Mode            string           `json:"mode"`
	Zone            zoneDTO          `json:"zone"`
	DeliveryAddress *string          `json:"deliveryAddress"`
	DeliveryFee     float64          `json:"deliveryFee"`
	Items           []orderItemDTO   `json:"items"`
	Subtotal        float64          `json:"subtotal"`
	Total           float64          `json:"total"`
	Reservation     reservationDTO   `json:"reservation"`
	Fulfillments    []fulfillmentDTO `json:"fulfillments"`
	Delivery        *deliveryDTO     `json:"delivery"`
	Payment         *paymentRefDTO   `json:"payment"`
	CreatedAt       string           `json:"createdAt"`
	PaidAt          *string          `json:"paidAt"`
}

func toOrderDTO(o domain.Order, now time.Time) orderDTO {
	d := orderDTO{
		ID: o.ID, Code: o.Code, Status: string(o.Status), ClientID: o.ClientID, ClientName: o.ClientName, Phone: o.Phone,
		PrescriptionID: o.PrescriptionID, Mode: string(o.Mode), Zone: zoneDTO(o.Zone), DeliveryAddress: strPtr(o.DeliveryAddress),
		DeliveryFee: o.DeliveryFee.Float(), Items: make([]orderItemDTO, 0, len(o.Items)),
		Subtotal: o.Subtotal().Float(), Total: o.Total().Float(),
		Reservation:  reservationDTO{ExpiresAt: ts(o.ReservationExpiresAt), SecondsLeft: o.SecondsLeft(now), Active: o.ReservationActive(now)},
		Fulfillments: make([]fulfillmentDTO, 0, len(o.Fulfillments)), CreatedAt: ts(o.CreatedAt), PaidAt: tsPtr(o.PaidAt),
	}
	for _, it := range o.Items {
		d.Items = append(d.Items, orderItemDTO{ID: it.ID, SKU: it.SKU, Medicine: it.Medicine, Brand: it.Brand,
			Presentation: it.Presentation, UnitLabel: it.UnitLabel, PharmacyID: it.PharmacyID, PharmacyName: it.PharmacyName,
			Quantity: it.Quantity, PrescribedQuantity: it.PrescribedQuantity, UnitPrice: it.UnitPrice.Float(),
			Subtotal: it.Subtotal().Float(), RequiresPrescription: it.RequiresPrescription,
			CanIncrease: it.CanIncrease(), CanDecrease: it.CanDecrease()})
	}
	for _, f := range o.Fulfillments {
		fd := fulfillmentDTO{PharmacyID: f.PharmacyID, PharmacyName: f.PharmacyName, Address: f.Address, Status: string(f.Status),
			ETA: tsPtr(f.ETA), ETAMinutes: f.ETAMinutes, Items: []fulfillmentItemDTO{}}
		for _, it := range o.ItemsFor(f.PharmacyID) {
			fd.Items = append(fd.Items, fulfillmentItemDTO{ID: it.ID, Medicine: it.Medicine, Brand: it.Brand, Quantity: it.Quantity})
		}
		d.Fulfillments = append(d.Fulfillments, fd)
	}
	if o.Delivery != nil {
		d.Delivery = &deliveryDTO{Status: string(o.Delivery.Status), ETA: tsPtr(o.Delivery.ETA), ETAMinutes: o.Delivery.ETAMinutes,
			Courier: courierDTO(o.Delivery.Courier)}
	}
	if o.Payment != nil {
		d.Payment = &paymentRefDTO{ID: o.Payment.ID, Method: string(o.Payment.Method), Status: string(o.Payment.Status)}
	}
	return d
}

func toOrderDTOs(os []domain.Order, now time.Time) []orderDTO {
	out := make([]orderDTO, 0, len(os))
	for _, o := range os {
		out = append(out, toOrderDTO(o, now))
	}
	return out
}

// --- Payments ---------------------------------------------------------------

type paymentDTO struct {
	ID          string  `json:"id"`
	OrderID     string  `json:"orderId"`
	Method      string  `json:"method"`
	Amount      float64 `json:"amount"`
	Status      string  `json:"status"`
	Link        *string `json:"link"`
	BankID      *string `json:"bankId"`
	CreatedAt   string  `json:"createdAt"`
	ConfirmedAt *string `json:"confirmedAt"`
}

func toPaymentDTO(p domain.Payment) paymentDTO {
	return paymentDTO{ID: p.ID, OrderID: p.OrderID, Method: string(p.Method), Amount: p.Amount.Float(), Status: string(p.Status),
		Link: strPtr(p.Link), BankID: strPtr(p.BankID), CreatedAt: ts(p.CreatedAt), ConfirmedAt: tsPtr(p.ConfirmedAt)}
}

type orderBriefDTO struct {
	ID     string  `json:"id"`
	Code   string  `json:"code"`
	Total  float64 `json:"total"`
	Status string  `json:"status"`
}

type bankDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type paymentOptionDTO struct {
	Method      string `json:"method"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// --- Misc -------------------------------------------------------------------

type notificationDTO struct {
	ID        string `json:"id"`
	Channel   string `json:"channel"`
	Recipient string `json:"recipient"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	OrderID   string `json:"orderId"`
	At        string `json:"at"`
}

type clientDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}
