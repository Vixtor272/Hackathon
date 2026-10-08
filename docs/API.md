# Farmi — REST contract (v1)

Base URL: `http://localhost:8080` (Go backend). Every endpoint lives under `/api/v1`.
The Svelte frontend runs on `http://localhost:5173` in dev (Vite proxies `/api` → `:8080`)
and is served by the Go server from `/` in production (SPA fallback).

Conventions
- JSON everywhere. Timestamps are RFC 3339 strings. Money is a JSON number in USD with 2 decimals (`12.5`).
- IDs are opaque strings (`ord_…`, `pay_…`, `rx_…`, `itm_…`, `msg_…`).
- Errors: `{ "error": { "code": "<CODE>", "message": "<texto en español para mostrar>" } }`

| HTTP | code |
|---|---|
| 400 | `VALIDATION` |
| 404 | `NOT_FOUND` |
| 409 | `INVALID_STATE`, `INSUFFICIENT_STOCK`, `RX_INCREASE_NOT_ALLOWED`, `RESERVATION_EXPIRED`, `EMPTY_CART`, `PAYMENT_INVALIDATED` |
| 422 | `OCR_ILLEGIBLE`, `PRESCRIPTION_INVALID` |
| 500 | `INTERNAL` |

---

## 1. WhatsApp channel (simulated)

The real WhatsApp Cloud API is replaced by a simulator with the same shape: inbound messages
arrive at a webhook, outbound messages go to a sender port. The UI reads the transcript.

### `POST /api/v1/whatsapp/webhook`
Inbound message from a client.
```json
{ "from": "+593991111111", "type": "text",  "text": "hola" }
{ "from": "+593991111111", "type": "image", "mediaId": "receta-001" }
```
Response `200`:
```json
{ "state": "ASK_PRESCRIPTION", "replies": [ Message, ... ] }
```

### `GET /api/v1/whatsapp/conversations/{phone}/messages`
```json
{ "phone": "+593991111111", "state": "ASK_ID", "orderId": null, "messages": [ Message, ... ] }
```
`Message`:
```json
{ "id": "msg_1", "direction": "in" | "out", "type": "text" | "image" | "link",
  "text": "¡Hola! Soy Farmi…", "mediaId": "receta-001", "mediaUrl": "/recetas/receta-001.svg",
  "link": "http://localhost:5173/checkout/ord_1", "at": "2026-10-08T15:04:05Z" }
```
`mediaId`/`mediaUrl` only on `image`; `link` only on `link` (the text still contains the URL).

### `DELETE /api/v1/whatsapp/conversations/{phone}` → `204`
Resets the conversation (same as the user typing `reiniciar`).

### `GET /api/v1/whatsapp/media`
Sample prescriptions the simulator can attach (what a phone camera would send).
```json
{ "media": [ { "id": "receta-001", "title": "Receta válida — Dra. Ana Torres",
               "description": "Paracetamol, Amoxicilina y Loratadina", "url": "/recetas/receta-001.svg",
               "expected": "valid" | "doctor_inactive" | "doctor_unknown" | "incomplete" | "illegible" } ] }
```

Conversation states (`state`): `ASK_ID`, `ASK_NAME`, `ASK_PRESCRIPTION`, `ASK_ZONE`, `ASK_MODE`,
`ASK_PICKUP_OPTION`, `ASK_ADDRESS`, `ASK_BRAND`, `CONFIRM_CART`, `AWAIT_PAYMENT`, `COMPLETED`.

Farmi understands numbers for list choices plus: `sí/no`, `retiro/domicilio`, `estado`, `cancelar`,
`continuar` (renew an expired reservation), `reiniciar`, `ayuda`.

---

## 2. OCR module

### `POST /api/v1/ocr/extract`
```json
{ "mediaId": "receta-001" }
```
`200`:
```json
{ "prescription": Prescription }
```
`422 OCR_ILLEGIBLE` when the image cannot be read.

`Prescription` (what the OCR returns, stored under `id`):
```json
{ "id": "rx_1", "mediaId": "receta-001",
  "patient": { "name": "María Pérez", "idNumber": "1712345678" },
  "issuedAt": "2026-10-07",
  "doctor": { "name": "Dra. Ana Torres", "registryId": "MSP-10234" },
  "hasSignature": true, "hasStamp": true, "confidence": 0.97,
  "items": [ { "medicine": "Paracetamol", "concentration": "500 mg", "presentation": "tabletas",
               "quantity": 20, "unit": "tabletas" } ] }
```

### `POST /api/v1/prescriptions/validate`
```json
{ "prescription": Prescription }
```
`200`:
```json
{ "valid": false, "errors": ["Falta la firma del médico"],
  "doctor": { "registryId": "MSP-10234", "name": "Dra. Ana Torres", "registered": true,
              "active": true, "enabledToPrescribe": true } }
```

---

## 3. Catalog & availability (AI module → company's external API, mocked)

### `GET /api/v1/catalog/zones`
```json
{ "zones": [ { "id": "uio-norte", "label": "Quito — zona norte" } ] }
```
### `GET /api/v1/catalog/pharmacies?zone=uio-norte`
```json
{ "pharmacies": [ { "id": "med-norte", "name": "Medicity Demo Norte", "chain": "Medicity",
                    "address": "Av. Ficticia A 123", "zone": "uio-norte" } ] }
```
### `POST /api/v1/catalog/availability`
```json
{ "zone": "uio-norte", "prescriptionId": "rx_1" }
```
`200`:
```json
{ "options": [
    { "id": "opt_1", "kind": "single" | "split", "label": "Medicity Demo Norte — toda la receta",
      "pharmacies": [ Pharmacy ],
      "coverage": [ { "medicine": "Paracetamol 500 mg", "pharmacyId": "med-norte", "quantity": 20 } ],
      "missing": [] } ],
  "missing": [ { "medicine": "Amoxicilina 500 mg", "requested": 21, "available": 0 } ],
  "deliveryAvailable": true }
```
### `POST /api/v1/catalog/brands`
```json
{ "pharmacyIds": ["med-norte"], "prescriptionId": "rx_1" }
```
`200`:
```json
{ "medicines": [ { "medicine": "Paracetamol 500 mg", "requested": 20, "requiresPrescription": false,
    "brands": [ { "sku": "PAR-ALFA-500", "brand": "Marca Alfa", "concentration": "500 mg",
                  "presentation": "tabletas", "sellByUnit": true, "unitsPerPack": 1,
                  "unitLabel": "tabletas", "unitPrice": 0.4, "quantity": 20, "subtotal": 8.0,
                  "pharmacyId": "med-norte", "available": 100 } ] } ] }
```
When `sellByUnit` is false the quantity is in boxes (`unitsPerPack` units each).

---

## 4. Orders (cart + reservation)

### `GET /api/v1/orders/{id}` → `{ "order": Order }`

`Order`:
```json
{ "id": "ord_1", "code": "DEMO-001",
  "status": "PENDING" | "PAID" | "PREPARING" | "READY" | "DISPATCHED" | "DELIVERED" | "CANCELLED" | "EXPIRED",
  "clientId": "1712345678", "clientName": "María Pérez", "phone": "+593991111111",
  "prescriptionId": "rx_1",
  "mode": "pickup" | "delivery",
  "zone": { "id": "uio-norte", "label": "Quito — zona norte" },
  "deliveryAddress": "Av. Demo 123" | null, "deliveryFee": 2.5,
  "items": [ { "id": "itm_1", "sku": "PAR-ALFA-500", "medicine": "Paracetamol 500 mg", "brand": "Marca Alfa",
               "presentation": "tabletas", "unitLabel": "tabletas",
               "pharmacyId": "med-norte", "pharmacyName": "Medicity Demo Norte",
               "quantity": 20, "prescribedQuantity": 20, "unitPrice": 0.4, "subtotal": 8.0,
               "requiresPrescription": false, "canIncrease": true, "canDecrease": true } ],
  "subtotal": 10.0, "total": 12.5,
  "reservation": { "expiresAt": "2026-10-08T15:14:05Z", "secondsLeft": 540, "active": true },
  "fulfillments": [ { "pharmacyId": "med-norte", "pharmacyName": "Medicity Demo Norte", "address": "…",
                      "status": "NOTIFIED" | "PREPARING" | "READY" | "PICKED_UP",
                      "eta": "2026-10-08T15:30:00Z" | null, "etaMinutes": 20,
                      "items": [ { "id": "itm_1", "medicine": "…", "brand": "…", "quantity": 20 } ] } ],
  "delivery": { "status": "CONSOLIDATING" | "DISPATCHED" | "DELIVERED", "eta": "…" | null, "etaMinutes": 45,
                "courier": { "name": "Carlos Repartidor (ficticio)", "phone": "+593990000000" } } | null,
  "payment": { "id": "pay_1", "method": "card" | "deuna", "status": "PENDING" | "APPROVED" | "REJECTED" | "INVALIDATED" } | null,
  "createdAt": "…", "paidAt": "…" | null }
```
Cart rules (enforced server-side, mirrored in `canIncrease`/`canDecrease`):
- Over-the-counter item: `+` and `−` allowed; `+` reserves extra stock or fails with `INSUFFICIENT_STOCK`.
- Prescription item: `−` only; `+` fails with `RX_INCREASE_NOT_ALLOWED`.
- Quantity `0` removes the item. Empty cart cannot be paid (`EMPTY_CART`).
- Any change recalculates totals, adjusts the reservation and invalidates a pending payment.

### `PATCH /api/v1/orders/{id}/items/{itemId}`
```json
{ "quantity": 15 }
```
→ `{ "order": Order }`

### `POST /api/v1/orders/{id}/cancel` → `{ "order": Order }` (releases reservations)
### `POST /api/v1/orders/{id}/renew` → `{ "order": Order }` (after `EXPIRED`: re-checks stock, new reservation)

### `GET /api/v1/orders/{id}/payment-options`
```json
{ "checkoutUrl": "http://localhost:5173/checkout/ord_1",
  "options": [ { "method": "card",  "label": "Tarjeta de débito o crédito", "description": "Simulador sin datos bancarios reales" },
               { "method": "deuna", "label": "DeUna",                        "description": "Paga con tu banco (simulado)" } ] }
```

---

## 5. Payments (simulators)

### `POST /api/v1/payments`
```json
{ "orderId": "ord_1", "method": "card" | "deuna" }
```
`201 { "payment": Payment }`. A new payment invalidates any previous `PENDING` payment of the order.
`409` if the order cannot be paid (`EMPTY_CART`, `RESERVATION_EXPIRED`, `INVALID_STATE`).

`Payment`:
```json
{ "id": "pay_1", "orderId": "ord_1", "method": "deuna", "amount": 12.5,
  "status": "PENDING" | "APPROVED" | "REJECTED" | "INVALIDATED",
  "link": "http://localhost:5173/deuna/pay_1" | null, "bankId": null,
  "createdAt": "…", "confirmedAt": null }
```
### `GET /api/v1/payments/{id}` → `{ "payment": Payment, "order": { "id", "code", "total", "status" } }`
### `GET /api/v1/payments/banks`
```json
{ "banks": [ { "id": "banco-demo-1", "name": "Banco Demo Pichincha" } ] }
```
### `POST /api/v1/payments/{id}/confirm`
```json
{ "outcome": "approved" | "rejected", "bankId": "banco-demo-1" }
```
→ `{ "payment": Payment, "order": Order }`. Idempotent: confirming an already-approved payment returns the
same result and never deducts stock twice. `409 PAYMENT_INVALIDATED` if the cart changed or the reservation expired.
On approval the order becomes `PAID`, stock is deducted once, reservations are consumed, pharmacies/courier are
notified and the client receives the WhatsApp confirmation with the ETA.

---

## 6. Fulfillment (cashier / courier board)

### `GET /api/v1/fulfillment/orders?pharmacyId=med-norte` | `?role=courier` | (none = all)
→ `{ "orders": [ Order ] }` — orders from `PAID` onwards.

### `POST /api/v1/fulfillment/orders/{id}/pharmacies/{pharmacyId}/status`
```json
{ "status": "PREPARING" | "READY" | "PICKED_UP" }
```
### `POST /api/v1/fulfillment/orders/{id}/delivery/status`
```json
{ "status": "DISPATCHED" | "DELIVERED" }
```
Both → `{ "order": Order }` and push the corresponding WhatsApp message with the ETA.

---

## 7. Notifications outbox & misc

### `GET /api/v1/notifications?channel=whatsapp|pharmacy|courier&orderId=ord_1`
```json
{ "notifications": [ { "id": "ntf_1", "channel": "pharmacy", "recipient": "med-norte",
                       "title": "Nuevo pedido DEMO-001", "body": "…", "orderId": "ord_1", "at": "…" } ] }
```
### `GET /api/v1/clients/{id}` → `{ "client": { "id": "1712345678", "name": "María Pérez", "phone": "+593991111111" } }`
### `GET /api/v1/health`
```json
{ "status": "ok", "modules": { "ocr": "mock", "catalogApi": "mock", "payments": "simulator",
                               "whatsapp": "simulator", "storage": "memory" } }
```

---

## Demo data

Clients (phone → cédula → name): `+593991111111` → `1712345678` María Pérez; `+593992222222` → `0912345678` Juan López.
Any other phone/cédula is registered on the fly (Farmi asks for the name).

Sample prescriptions (`mediaId`):
| id | result | contents |
|---|---|---|
| `receta-001` | valid | Dra. Ana Torres (MSP-10234). Paracetamol 500 mg ×20, Amoxicilina 500 mg ×21, Loratadina 10 mg ×10 |
| `receta-002` | valid | Dr. Luis Andrade (MSP-20987). Ibuprofeno 400 mg ×12, Omeprazol 20 mg ×14, Metformina 850 mg ×30 |
| `receta-003` | doctor inactive | Dr. Pedro Salazar (MSP-30111) is registered but inactive |
| `receta-004` | incomplete | no signature, no stamp |
| `receta-005` | illegible | OCR cannot read the image |
| `receta-006` | doctor unknown | registry id `MSP-99999` does not exist |

Zones: `uio-norte` Quito — zona norte · `uio-centro` Quito — zona centro · `uio-sur` Quito — zona sur · `gye-norte` Guayaquil — zona norte.

Pharmacies: `med-norte` Medicity Demo Norte · `eco-norte` Económicas Demo Norte (uio-norte) ·
`eco-centro` Económicas Demo Centro · `med-centro` Medicity Demo Centro (uio-centro) · `med-sur` Medicity Demo Sur (uio-sur) ·
`eco-gye` Económicas Demo Guayaquil (gye-norte).

Stock is arranged so that, for `receta-001`: zona norte → one pharmacy covers everything (`med-norte`);
zona centro → split between `eco-centro` + `med-centro`; zona sur → missing items for pickup but delivery works
(company stock = sum of all pharmacies). Delivery fee USD 2.50. Reservation TTL 10 min.
ETAs: pickup ready ≈ 20 min · delivery arrival ≈ 45 min · after dispatch ≈ 15 min.
