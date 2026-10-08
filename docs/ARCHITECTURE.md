# Farmi — architecture and flow

Diagrams of the MVP. They render on GitHub and in any Mermaid-aware viewer. The REST contract is in
[`API.md`](API.md) and the business flow it implements is in [`../flujo_farmaenlace.md`](../flujo_farmaenlace.md).

## 1. Two apps, two API surfaces, one core

The customer experience and the company's back office are separate apps on separate ports. The Go backend is
one process with one set of data, but each port mounts only its half of the API: the client's browser cannot
reach the cashier board, and the portal cannot talk to the WhatsApp channel.

```mermaid
flowchart LR
    subgraph CLIENT["Client app · Vite :5173 / built on :8080"]
        WA["WhatsApp simulator<br/>/"]
        CO["Checkout + payment<br/>/checkout/:id"]
        DU["DeUna simulator<br/>/deuna/:id"]
    end

    subgraph COMPANY["Company portal · Vite :5174 / built on :8081"]
        CB["Cashier boards<br/>one tab per pharmacy"]
        RB["Courier board"]
        NO["Notifications outbox"]
    end

    subgraph GO["Go backend · one process"]
        CS["Customer surface :8080<br/>whatsapp · ocr · catalog · orders · payments"]
        KS["Company surface :8081<br/>fulfillment · notifications · clients<br/>read-only orders and pharmacies"]
        subgraph CORE["Core (domain + services, no I/O)"]
            FA["farmi · chat state machine"]
            OC["ocr + prescription validator"]
            CA["catalog · AI availability"]
            OR["order · cart + 10-min hold"]
            PA["payment · idempotent"]
            FU["fulfillment · ETAs"]
            NT["notify"]
        end
        EW["expiry worker · every 5 s"]
    end

    subgraph ADAPTERS["Simulated adapters (driven ports)"]
        A1["ocr/mock"]
        A2["doctors/memory"]
        A3["catalogapi/mock"]
        A4["ai/rules"]
        A5["payment/simulator"]
        A6["whatsapp/simulator"]
        A7["storage/memory"]
    end

    CLIENT -- "/api/v1" --> CS
    COMPANY -- "/api/v1" --> KS
    CS -- "driving ports" --> CORE
    KS -- "driving ports" --> CORE
    EW -- "releases lapsed holds" --> OR
    CORE -- "driven ports" --> ADAPTERS
```

Swapping a simulator for the real system means writing one adapter and changing one line in
`backend/cmd/server/main.go`; the core, the chat flow and both apps stay as they are.

## 2. A pickup purchase, step by step

```mermaid
sequenceDiagram
    autonumber
    actor C as Client
    participant F as Farmi (WhatsApp)
    participant K as Core (catalog · order)
    participant W as Checkout page
    participant P as Card / DeUna simulator
    participant O as Cashier (company portal)

    rect rgba(13, 119, 107, 0.08)
    Note over C,K: 1 · Identify and validate
    C->>F: hola
    F-->>C: asks for the cédula first
    C->>F: cédula
    C->>F: prescription photo
    F->>K: OCR → JSON, doctor registry check
    K-->>F: valid, or the reasons it is not
    end

    rect rgba(13, 119, 107, 0.04)
    Note over C,K: 2 · Choose where and what
    C->>F: zone
    F->>K: availability(zone, prescription)
    K-->>F: stores with everything (only those), else nearest pair
    C->>F: retiro / domicilio
    Note right of F: one option → selected without asking
    C->>F: brand per medicine · sí
    F->>K: create order, hold stock 10 min
    F-->>C: checkout link
    end

    rect rgba(161, 98, 7, 0.08)
    Note over C,P: 3 · Confirm and pay (within the 10-minute hold)
    C->>W: opens the link
    Note over W: detects computer or phone
    W->>K: − / + quantities (prescription items ≤ prescribed)
    W->>P: card form, or DeUna
    Note over P: computer → QR with random reference<br/>phone → DeUna page, pick a bank
    P->>K: approved
    Note over K: PAID · stock deducted once
    end

    rect rgba(54, 81, 166, 0.08)
    Note over K,O: 4 · Prepare and hand over
    K->>O: new order notification
    K-->>C: paid + ETA ≈ 20 min (WhatsApp)
    O->>K: En preparación → Listo para recoger
    K-->>C: listo para recoger (WhatsApp)
    C-->>O: picks up with the order code
    end
```

For a home delivery, step 2 asks for the address instead of a pharmacy, the company chooses the origin stores,
and in step 4 the courier board moves the order to **En reparto** (ETA ≈ 15 min) and **Entregado**.

## 3. Order states

```mermaid
stateDiagram-v2
    direction LR
    [*] --> PENDING: order created, stock held
    PENDING --> PAID: payment approved (stock deducted once)
    PENDING --> EXPIRED: 10 minutes pass (units released)
    EXPIRED --> PENDING: continuar (stock re-reserved)
    PENDING --> CANCELLED: cancelar (units released)
    EXPIRED --> CANCELLED: cancelar
    state "Pickup" as pickup {
        PREPARING --> READY: every pharmacy ready
    }
    PAID --> PREPARING: cashier starts
    READY --> DELIVERED: picked up
    PAID --> DISPATCHED: courier leaves (delivery)
    DISPATCHED --> DELIVERED: delivered
    DELIVERED --> [*]
    CANCELLED --> [*]
```

## 4. Checkout rules

**Which pharmacy Farmi suggests**

```mermaid
flowchart LR
    S["Stores in the zone<br/>available = stock − reservations"] --> Q{"Does one store<br/>have it all?"}
    Q -- yes --> A["Suggest only those stores<br/>(one → no question)"]
    Q -- no --> B["Main store + the nearest one<br/>that has the rest"]
    Q -. nobody .-> M["List what is missing<br/>(delivery may still work)"]
```

**Prescription quantities.** A prescription item prescribed at 21 can go down to any amount and back up to 21,
never above (`RX_INCREASE_NOT_ALLOWED`). Zero removes the line, and the page asks before removing the last unit.
Over-the-counter items can grow as far as stock allows.

**DeUna by device**

```mermaid
flowchart LR
    D{"Device?<br/>user agent + pointer"} -- computer --> QR["QR of the DeUna link<br/>+ random reference DU-XXXXXXXX"]
    QR --> SC["Scan with the phone<br/>(simulated) → PAID"]
    D -- phone --> DP["Open /deuna/:id"] --> BK["Pick a bank → confirm → PAID"]
```

The client can switch modes by hand ("Estoy en el celular" / "Estoy en un computador"). A cart change, a
rejection or an expiry voids the QR, and the next one carries a new reference.

**Card form.** Number with live grouping, brand detection (Visa, Mastercard, Amex, Diners, Discover) and the
Luhn check digit; cardholder; `MM/AA` expiry not in the past; CVV of 3 digits (4 on Amex). Errors show under each
field. Card data never leaves the browser: the sandbox number decides the outcome
(`4000 0000 0000 0002` → declined, any other valid number → approved).
