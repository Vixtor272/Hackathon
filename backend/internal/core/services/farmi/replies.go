package farmi

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/farmaenlace/farmi/internal/core/domain"
	"github.com/farmaenlace/farmi/internal/core/ports"
)

// Every string the client reads lives here, so the flow in assistant.go stays
// readable and the copy can be tuned without touching logic.
const (
	msgGreeting             = "¡Hola! Soy *Farmi* 💊, el asistente de compras de Farmaenlace. Te ayudo a comprar los medicamentos de tu receta en Medicity y Farmacias Económicas.\n\nAntes de empezar, ¿me indicas tu número de cédula?"
	msgGreetingAgain        = "¡Hola de nuevo! 👋 Para iniciar una nueva compra, indícame tu número de cédula."
	msgInvalidID            = "Necesito tu número de cédula (10 dígitos) para continuar. Por ejemplo: 1712345678."
	msgIDBeforePrescription = "Antes de revisar tu receta necesito tu número de cédula (10 dígitos)."
	msgAskNameAgain         = "¿Me indicas tu nombre para registrarte?"
	msgNeedPhoto            = "📷 Por favor envíame una foto de tu receta para continuar."
	msgIllegible            = "No pude leer la receta: la imagen es ilegible o no corresponde a una receta. Por favor envía otra foto más clara."
	msgAskMode              = "¿Prefieres retirar en farmacia o recibir tu pedido a domicilio? Responde *retiro* o *domicilio*."
	msgModeNotUnderstood    = "No te entendí. Responde *retiro* para recoger en farmacia o *domicilio* para recibirlo en casa."
	msgNoPickup             = "No hay opciones de retiro en esta zona que cubran tu receta. Puedes elegir *domicilio* o escribir *reiniciar* para cambiar de zona."
	msgNoDelivery           = "Por ahora no tenemos stock de toda tu receta para entrega a domicilio. Elige *retiro* o escribe *reiniciar* para cambiar de zona."
	msgAskAddress           = "🏠 Escribe la dirección de entrega (calle, número y una referencia)."
	msgAskAddressAgain      = "Necesito una dirección de entrega válida (al menos calle y número)."
	msgConfirmNotUnderstood = "Responde *sí* para confirmar la compra o *no* para elegir otras marcas."
	msgRechooseBrands       = "Entendido, elijamos de nuevo las marcas."
	msgAfterCancel          = "Cuando quieras volver a comprar, escribe *nueva compra*."
	msgNewPurchaseHint      = "Escribe *nueva compra* para iniciar otra compra."
	msgExpiredHint          = "Escribe *continuar* para verificar disponibilidad y generar una nueva reserva."
	msgInternalError        = "Ocurrió un problema interno. Por favor intenta de nuevo en unos segundos."
)

func msgAskName(id string) string {
	return fmt.Sprintf("No encontré un registro con la cédula %s. ¿Cómo te llamas?", id)
}

func msgAskPrescription(name string) string {
	return fmt.Sprintf("Gracias, %s. 📷 Envíame una foto de tu receta y te ayudo a encontrar tus medicamentos.", name)
}

func msgInvalidPrescription(errs []string) string {
	return "No pude validar la receta por estos motivos:\n" + bullets(errs) + "\n\nPor favor envía otra receta."
}

func msgPrescriptionOK(rx domain.Prescription, res domain.ValidationResult) string {
	var b strings.Builder
	b.WriteString("✅ Receta validada.\n")
	fmt.Fprintf(&b, "👤 Paciente: %s\n", rx.Patient.Name)
	fmt.Fprintf(&b, "🩺 Médico: %s (%s) — registrado, activo y habilitado\n", res.Doctor.Name, res.Doctor.RegistryID)
	fmt.Fprintf(&b, "📅 Emitida: %s\n\n💊 Medicamentos:\n", rx.IssuedAt)
	for _, it := range rx.Items {
		fmt.Fprintf(&b, "• %s — %d %s\n", it.Label(), it.Quantity, it.Unit)
	}
	return strings.TrimRight(b.String(), "\n")
}

func msgAskZone(zones []domain.Zone) string {
	var b strings.Builder
	b.WriteString("¿En qué ciudad o zona deseas comprar?\n")
	for i, z := range zones {
		fmt.Fprintf(&b, "%d. %s\n", i+1, z.Label)
	}
	b.WriteString("\nResponde con el número.")
	return b.String()
}

func msgZoneNotUnderstood(zones []domain.Zone) string {
	return "No identifiqué la zona. " + msgAskZone(zones)
}

func msgNoAvailability(zone domain.Zone, missing []domain.MissingItem) string {
	return fmt.Sprintf("En %s no encontré disponibilidad para tu receta:\n%s\n\nElige otra zona.", zone.Label, bulletMissing(missing))
}

func msgOptions(av domain.Availability) string {
	var b strings.Builder
	if len(av.Options) == 0 {
		fmt.Fprintf(&b, "No hay farmacias en %s que cubran tu receta para retiro.\n", av.Zone.Label)
	} else {
		fmt.Fprintf(&b, "Estas son las opciones en %s:\n\n", av.Zone.Label)
		for i, opt := range av.Options {
			fmt.Fprintf(&b, "%d. 🏪 %s\n", i+1, opt.Label)
			if opt.Kind == domain.OptionSingle {
				fmt.Fprintf(&b, "   📍 %s\n", opt.Pharmacies[0].Address)
			} else {
				for _, ph := range opt.Pharmacies {
					fmt.Fprintf(&b, "   📍 %s (%s): %s\n", ph.Name, ph.Address, strings.Join(medicinesAt(opt, ph.ID), ", "))
				}
			}
			b.WriteString("\n")
		}
	}
	if len(av.Missing) > 0 {
		fmt.Fprintf(&b, "⚠️ Para retiro en esta zona no hay disponibilidad de:\n%s\n", bulletMissing(av.Missing))
	}
	if av.DeliveryAvailable {
		b.WriteString("🚚 A domicilio sí podemos completar toda tu receta.")
	} else {
		b.WriteString("🚚 Por ahora no es posible completar toda tu receta a domicilio.")
	}
	return strings.TrimRight(b.String(), "\n")
}

func medicinesAt(opt domain.PharmacyOption, pharmacyID string) []string {
	var out []string
	for _, c := range opt.Coverage {
		if c.PharmacyID == pharmacyID {
			out = append(out, c.Medicine)
		}
	}
	return out
}

func msgAskPickupOption(n int) string {
	if n == 1 {
		return "¿Confirmas que retirarás en la opción 1? Responde con el número 1."
	}
	return fmt.Sprintf("¿En cuál opción deseas retirar? Responde con el número (1 a %d).", n)
}

func msgOptionChosen(opt domain.PharmacyOption) string {
	var b strings.Builder
	fmt.Fprintf(&b, "👍 Elegiste: %s.", opt.Label)
	if opt.Kind == domain.OptionSplit {
		b.WriteString("\nRecogerás en cada local lo siguiente:")
		for _, ph := range opt.Pharmacies {
			fmt.Fprintf(&b, "\n• %s: %s", ph.Name, strings.Join(medicinesAt(opt, ph.ID), ", "))
		}
	}
	return b.String()
}

func msgNoDeliveryNow(missing []domain.MissingItem) string {
	return "En este momento no tenemos stock para entregar toda tu receta a domicilio:\n" + bulletMissing(missing)
}

func msgDeliveryAssigned(addr string, c domain.Courier, fee domain.Money) string {
	return fmt.Sprintf("✅ Entrega a domicilio en: %s\n🛵 Repartidor asignado: %s\n💵 Costo de envío: %s\n\nNosotros nos encargamos de obtener los productos; no necesitas elegir farmacias.",
		addr, c.Name, fee.Format())
}

func msgNoBrands(medicine string) string {
	return fmt.Sprintf("No encontré marcas disponibles para %s en los locales elegidos. Elijamos otra zona.", medicine)
}

func msgBrandPrompt(mb domain.MedicineBrands) string {
	var b strings.Builder
	fmt.Fprintf(&b, "💊 *%s* — necesitas %d %s (%s).\nElige una marca:\n", mb.Medicine, mb.Requested, mb.Unit, conditionLabel(mb.RequiresPrescription))
	for i, br := range mb.Brands {
		fmt.Fprintf(&b, "%d. %s — %s %s — %s × %s = %s (%s)\n", i+1, br.Product.Brand, br.Product.Concentration,
			br.Product.Presentation, quantityLabel(br), br.Product.UnitPrice.Format(), br.Subtotal().Format(), br.PharmacyName)
	}
	b.WriteString("\nResponde con el número.")
	return b.String()
}

func msgChooseNumber(n int) string {
	return fmt.Sprintf("Responde con un número entre 1 y %d.", n)
}

func msgCart(conv *domain.Conversation) string {
	var b strings.Builder
	b.WriteString("🛒 Tu carrito:\n")
	for _, s := range conv.Selections {
		fmt.Fprintf(&b, "• %s — %s — %s — %s × %s = %s\n", s.Medicine, s.Product.Brand,
			conditionLabel(s.Product.RequiresPrescription), quantityLabel(s), s.Product.UnitPrice.Format(), s.Subtotal().Format())
	}
	subtotal := conv.CartSubtotal()
	fmt.Fprintf(&b, "\nTotal de productos: %s", subtotal.Format())
	if conv.Mode == domain.ModeDelivery {
		fmt.Fprintf(&b, "\n🚚 Envío: %s\n💰 Total a pagar: %s", domain.DefaultDeliveryFee.Format(), (subtotal + domain.DefaultDeliveryFee).Format())
	} else {
		fmt.Fprintf(&b, "\n💰 Total a pagar: %s", subtotal.Format())
	}
	b.WriteString("\n\n¿Confirmas la compra? Responde *sí* o *no*.")
	return b.String()
}

func msgReservationFailed(reason string) string {
	return fmt.Sprintf("⚠️ No pude reservar todo el carrito: %s.\nVamos a ajustar las opciones con el stock actual.", reason)
}

func msgReserved(o domain.Order, now time.Time) string {
	return fmt.Sprintf("✅ Pedido *%s* creado. Reservé tus productos por %d minutos.", o.Code, minutesUntil(o.ReservationExpiresAt, now))
}

func msgPayLink(opts ports.PaymentOptions) string {
	labels := make([]string, 0, len(opts.Options))
	for _, o := range opts.Options {
		labels = append(labels, o.Label)
	}
	return fmt.Sprintf("💳 Confirma tu carrito y paga aquí (%s):\n%s", strings.Join(labels, " o "), opts.CheckoutURL)
}

func msgStatus(o domain.Order, now time.Time) string {
	switch o.Status {
	case domain.OrderPending:
		if o.ReservationActive(now) {
			return fmt.Sprintf("📋 Pedido %s: pendiente de pago. Reserva vigente por %d min.", o.Code, minutesUntil(o.ReservationExpiresAt, now))
		}
		return fmt.Sprintf("📋 Pedido %s: pendiente de pago, pero la reserva venció.", o.Code)
	case domain.OrderExpired:
		return fmt.Sprintf("⏰ Pedido %s: la reserva venció y las unidades fueron liberadas.", o.Code)
	case domain.OrderCancelled:
		return fmt.Sprintf("❌ Pedido %s: cancelado.", o.Code)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "📋 Pedido %s: pagado ✅\n", o.Code)
	if o.Mode == domain.ModeDelivery && o.Delivery != nil {
		switch o.Delivery.Status {
		case domain.DeliveryDispatched:
			fmt.Fprintf(&b, "🛵 En reparto con %s. Llegada en aprox. %d min a %s.", o.Delivery.Courier.Name, minutesUntilPtr(o.Delivery.ETA, now), o.DeliveryAddress)
		case domain.DeliveryDelivered:
			b.WriteString("🏠 Entregado. ¡Gracias por tu compra!")
		default:
			fmt.Fprintf(&b, "📦 Preparando tu pedido para el envío a %s. Llegada en aprox. %d min.", o.DeliveryAddress, minutesUntilPtr(o.Delivery.ETA, now))
		}
		return b.String()
	}
	for _, f := range o.Fulfillments {
		fmt.Fprintf(&b, "• %s (%s): %s\n", f.PharmacyName, f.Address, fulfillmentLabel(f, now))
	}
	return strings.TrimRight(b.String(), "\n")
}

func fulfillmentLabel(f domain.Fulfillment, now time.Time) string {
	switch f.Status {
	case domain.FulfillmentPreparing:
		return fmt.Sprintf("en preparación, listo en aprox. %d min", minutesUntilPtr(f.ETA, now))
	case domain.FulfillmentReady:
		return "listo para recoger ✅"
	case domain.FulfillmentPickedUp:
		return "entregado ✅"
	}
	return fmt.Sprintf("en espera de preparación, listo en aprox. %d min", minutesUntilPtr(f.ETA, now))
}

func msgRenewFailed(reason string) string {
	return fmt.Sprintf("No pude generar una nueva reserva: %s.", reason)
}

func msgRenewed(o domain.Order) string {
	return fmt.Sprintf("🔄 Generé una nueva reserva de 10 minutos para tu pedido %s.", o.Code)
}

func helpFor(state domain.ConversationState) string {
	steps := map[domain.ConversationState]string{
		domain.StateAskID:           "indicar tu número de cédula",
		domain.StateAskName:         "indicar tu nombre",
		domain.StateAskPrescription: "enviar la foto de tu receta",
		domain.StateAskZone:         "elegir la ciudad o zona",
		domain.StateAskMode:         "elegir retiro o domicilio",
		domain.StateAskPickupOption: "elegir la farmacia",
		domain.StateAskAddress:      "escribir la dirección de entrega",
		domain.StateAskBrand:        "elegir la marca de cada medicamento",
		domain.StateConfirmCart:     "confirmar el carrito",
		domain.StateAwaitPayment:    "pagar desde el enlace",
		domain.StateCompleted:       "iniciar una nueva compra",
	}
	return "Soy Farmi 💊. Comandos: *estado* (ver tu pedido), *cancelar* (cancelar el pedido pendiente), *continuar* (renovar la reserva), *reiniciar* (empezar de nuevo).\nPaso actual: " + steps[state] + "."
}

func conditionLabel(requiresRx bool) string {
	if requiresRx {
		return "Bajo receta"
	}
	return "Venta libre"
}

func quantityLabel(b domain.BrandOption) string {
	if b.Product.SellByUnit || b.Product.UnitsPerPack <= 1 {
		return fmt.Sprintf("%d %s", b.Quantity, b.Product.UnitLabel)
	}
	return fmt.Sprintf("%d %s de %d", b.Quantity, b.Product.UnitLabel, b.Product.UnitsPerPack)
}

func bullets(lines []string) string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, "• "+l)
	}
	return strings.Join(out, "\n")
}

func bulletMissing(missing []domain.MissingItem) string {
	lines := make([]string, 0, len(missing))
	for _, m := range missing {
		lines = append(lines, fmt.Sprintf("%s: necesitas %d, disponibles %d", m.Medicine, m.Requested, m.Available))
	}
	return bullets(lines)
}

func minutesUntil(t, now time.Time) int {
	m := int(math.Ceil(t.Sub(now).Minutes()))
	if m < 1 {
		return 1
	}
	return m
}

func minutesUntilPtr(t *time.Time, now time.Time) int {
	if t == nil {
		return 0
	}
	return minutesUntil(*t, now)
}
