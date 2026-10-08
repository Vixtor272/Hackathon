# Farmi — asistente de compras por WhatsApp (MVP, hackathon Farmaenlace)

Farmi lets a client buy the medicines of a prescription from WhatsApp. This repository is a
**functional MVP** of the flow described in `flujo_farmaenlace.md`:

1. Farmi greets and **always asks for the client's cédula** before anything else.
2. The client sends the prescription photo → **OCR** returns structured JSON → the prescription is
   **validated** (required fields + doctor registered, active and enabled to prescribe in Ecuador).
3. The client picks a city/zone → the **AI module consults the company's product API** and returns
   the options: one pharmacy with everything, a main pharmacy + the nearest one with the rest, or
   what is missing.
4. Pickup (choose the pharmacy/combination) or home delivery (company-wide stock, courier assigned,
   shipping fee).
5. Brand and price per medicine → cart with totals.
6. Stock is re-checked and **reserved for 10 minutes**; Farmi sends the **payment link**.
7. The web page shows the cart (OTC items can grow if there is stock, prescription items can only
   shrink), the countdown and the **payment options**: card simulator or DeUna simulator.
8. On approval the payment is registered, the order becomes PAID, stock is deducted **once**
   (idempotent), pharmacies/courier are notified and the client receives the **ETA** by WhatsApp,
   for pickup and for delivery. Cashiers and couriers advance the order from an operations board
   and every step notifies the client.
9. Expired reservations release stock (`continuar` renews them); `cancelar` releases them too.

Everything external is **simulated behind interfaces**: WhatsApp, OCR, the medical registry, the
catalog/inventory API, the payment rails and the storage. No real API is called and nothing is
charged.

## Stack

| Layer    | Choice                                                                                                                                                              |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Backend  | Go 1.27, standard library only, REST API (`docs/API.md`), hexagonal architecture                                                                                  |
| Frontend | Svelte 5 + TypeScript + Vite, hexagonal architecture with ports/adapters/use cases, no UI framework                                                                 |
| Tests    | `go test` (domain rules + use cases through the hexagon), vitest (use cases with in-memory gateways), `scripts/e2e_smoke.py` (59 checks against a live backend) |

## Run it

Prerequisites: Go ≥ 1.22 and Node ≥ 20. On this machine Go 1.27.2 is installed in `~/sdk/go1.27.2`
and linked from `~/bin` (already on the PATH in `~/.zshrc`).

```bash
make install   # npm install
make run       # builds the Svelte app and serves app + API from Go on http://localhost:8080
```

Development with hot reload (two terminals):

```bash
make backend   # Go API on :8080, payment links point to :5173
make frontend  # Vite on :5173, proxies /api → :8080
```

Checks:

```bash
make test      # go vet + go test, svelte-check, vitest
make smoke     # starts a backend on :18080 and drives the whole flow through the REST contract
```

Configuration (environment variables of the backend): `PORT` (8080), `WEB_BASE_URL` (base of the
links sent by WhatsApp), `STATIC_DIR` (built frontend to serve, `none` for API only; defaults to
`../frontend/dist` when it exists), `RESERVATION_TTL` (10m), `EXPIRY_TICK` (5s).

## Demo script (≈5 minutes)

Pages: `/` WhatsApp simulator · `/checkout/:orderId` cart + payment · `/deuna/:paymentId` bank
simulator · `/operaciones` cashier / courier board + notifications outbox.

1. **Identity.** In `/` pick María (+593991111111), type `hola`. Farmi asks for the cédula → `1712345678`.
2. **OCR + validation.** Attach `receta-004` (no signature/stamp) → rejected with the reasons.
   Attach `receta-003` (inactive doctor) → rejected. Attach `receta-001` → validated, medicines listed.
3. **Availability.** Zone `2` (Quito — zona centro) → the only option is the split
   "Económicas Demo Centro + Medicity Demo Centro" (zone `1` gives a single pharmacy, zone `3` has
   nothing for pickup but delivery works). `retiro` → `1`.
4. **Brands and cart.** Pick a brand per medicine (`1`, `1`, `1`), confirm with `sí` → order
   `DEMO-001`, 10-minute reservation, checkout link.
5. **Checkout.** Open the link: `+`/`−` per item (prescription items only `−`), countdown,
   DeUna (pick a bank, reject, then retry) or card ("Simular pago aprobado").
6. **ETA + fulfilment.** The chat receives the confirmation with the ETA per pharmacy. In
   `/operaciones` open each pharmacy tab → "En preparación" → "Listo para recoger": the client is
   notified at every step; `estado` in the chat shows the same.
7. **Delivery.** Juan (+593992222222) → `0912345678` → `receta-002` → `guayaquil` → `domicilio` →
   an address → courier + USD 2,50 → brands → `sí` → pay → `/operaciones` → Reparto → "En reparto"
   → "Entregado".
8. **Expiry.** Run the backend with `RESERVATION_TTL=1m` to see the reservation expire; `continuar`
   renews it, `cancelar` releases it.

Sample data (clients, prescriptions, zones, pharmacies, stock) is listed in `docs/API.md` → "Demo data".

## Architecture

Both halves are hexagons: the core knows nothing about HTTP, WhatsApp, storage or the DOM; it talks
to the outside only through interfaces (ports) that adapters implement. Swapping a simulator for the
real system means writing one adapter and changing one line in the composition root.

```
                      WhatsApp (simulated)      Svelte app (checkout, DeUna, operations)
                              │                                   │
                              ▼                                   ▼
                  ┌──────────────────────── REST adapter (Go net/http) ───────────────────────┐
                  │                      internal/adapters/in/rest                            │
                  └───────────────┬───────────────────────────────────────────┬──────────────┘
                                  ▼           driving ports (core/ports)      ▼
  ┌──────────────────────────────────────────── core ─────────────────────────────────────────┐
  │  farmi (chat state machine)  ocr  prescription (validator)  catalog (AI / availability)   │
  │  order (cart + reservation)  payment (rails, idempotent)  fulfillment (ETA)  notify       │
  │                              domain: entities + business rules                            │
  └───────────────┬───────────────────────────────────────────────────────────┬──────────────┘
                  ▼           driven ports (core/ports/driven.go)             ▼
   ocr/mock · doctors/memory · catalogapi/mock · ai/rules · payment/simulator · whatsapp/simulator
   storage/memory (clients, prescriptions, orders, payments, conversations, messages, notifications, inventory)
```

### Backend (`backend/`)

```
cmd/server/main.go                 composition root: builds adapters, injects them into services, starts HTTP
internal/core/domain/              entities and rules: Order (cart rules, totals, ETAs), Prescription, Payment, Conversation…
internal/core/ports/               driving.go (what the app offers) · driven.go (what it needs)
internal/core/services/
  farmi/        Farmi: conversation state machine (ASK_ID → … → AWAIT_PAYMENT) + Spanish replies
  ocr/          extraction use case (image → JSON)
  prescription/ validation (required fields + doctor registry)
  catalog/      AI module: medicine matching, pharmacy options (single/split/missing), delivery plan, brands
  order/        cart, 10-minute reservation, web-page quantity rules, cancel, expiry, renew
  payment/      card/DeUna attempts, idempotent approval, stock deduction, notifications
  fulfillment/  cashier / courier steps, ETA messages
  notify/       fan-out to WhatsApp + outbox
internal/adapters/in/rest/         handlers, DTOs (contract in docs/API.md), SPA static serving
internal/adapters/out/             one package per simulated system (see diagram)
internal/demo/                     fictitious dataset
internal/testkit/                  assembles the hexagon with in-memory adapters + fake clock for tests
```

### Frontend (`frontend/src/lib`)

```
domain/          types (Order, Payment, Message…) and labels
application/
  ports/         ChatGateway, OrderGateway, PaymentGateway, FulfillmentGateway, CatalogGateway, NotificationGateway
  usecases/      ChatSession, Checkout, PaymentFlow, OperationsBoard (depend only on ports)
adapters/http/   fetch implementations of the ports (base /api/v1, error envelope → ApiError)
adapters/memory/ in-memory fakes used by vitest
container.ts     composition root · router.ts tiny history router
ui/              Svelte components per page: whatsapp/, checkout/, deuna/, operations/
```

## Requested modules ↔ code

| Module                               | Where                                                                                                                                                           |
| ------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Farmi chatbot on WhatsApp            | `services/farmi`, `adapters/out/whatsapp/simulator`, `frontend/.../ui/whatsapp`                                                                           |
| OCR + prescription validation        | `services/ocr`, `services/prescription`, `adapters/out/ocr/mock`, `adapters/out/doctors/memory`                                                         |
| AI module consulting the company API | `services/catalog` (availability, delivery plan, brands) + `adapters/out/ai/rules` (intents, zone and medicine matching) + `adapters/out/catalogapi/mock` |
| Cart availability & reservation      | `services/order` + `adapters/out/storage/memory/inventory.go` (availability = stock − reservations)                                                        |
| Payment module                       | `services/payment`, `adapters/out/payment/simulator`, `frontend/.../ui/checkout`, `ui/deuna`                                                            |
| Web server                           | `adapters/in/rest` (REST + serves the SPA), `cmd/server`                                                                                                    |
| Data storage                         | ports in`core/ports/driven.go`; `adapters/out/storage/memory` (8 repositories)                                                                              |
| Fulfilment & notifications (ETA)     | `services/fulfillment`, `services/notify`, `frontend/.../ui/operations`                                                                                   |

## Going from simulated to real

| Simulated today                                    | Real adapter to write                                 | Port                                    |
| -------------------------------------------------- | ----------------------------------------------------- | --------------------------------------- |
| `whatsapp/simulator` (transcript in memory)      | Meta WhatsApp Cloud API webhook + send                | `MessageSender`, `MessageLog`       |
| `ocr/mock` (preloaded JSON per image)            | Vision/OCR service returning the same`Prescription` | `PrescriptionExtractor`               |
| `doctors/memory`                                 | Ministry / company registry API                       | `DoctorRegistry`                      |
| `catalogapi/mock` + `storage/memory/inventory` | Farmaenlace product & stock API                       | `CatalogAPI`, `InventoryRepository` |
| `ai/rules` (keywords + fuzzy matching)           | LLM-backed interpreter (e.g. Claude)                  | `ConversationAI`                      |
| `payment/simulator`                              | Card processor / DeUna                                | `PaymentGateway`                      |
| `storage/memory`                                 | PostgreSQL                                            | the eight repository ports              |

The core, the chat flow and the frontend do not change.
