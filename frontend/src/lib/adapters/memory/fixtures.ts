import type { Order, OrderItem } from '../../domain';

/** Deterministic sample order used by the in-memory fakes and the use-case tests. */
export function sampleOrder(overrides: Partial<Order> = {}): Order {
  const items: OrderItem[] = [
    {
      id: 'itm_1',
      sku: 'PAR-ALFA-500',
      medicine: 'Paracetamol 500 mg',
      brand: 'Marca Alfa',
      presentation: 'tabletas',
      unitLabel: 'tabletas',
      pharmacyId: 'med-norte',
      pharmacyName: 'Medicity Demo Norte',
      quantity: 20,
      prescribedQuantity: 20,
      unitPrice: 0.4,
      subtotal: 8,
      requiresPrescription: false,
      canIncrease: true,
      canDecrease: true,
    },
    {
      id: 'itm_2',
      sku: 'AMX-BETA-500',
      medicine: 'Amoxicilina 500 mg',
      brand: 'Marca Beta',
      presentation: 'cápsulas',
      unitLabel: 'cápsulas',
      pharmacyId: 'med-norte',
      pharmacyName: 'Medicity Demo Norte',
      quantity: 21,
      prescribedQuantity: 21,
      unitPrice: 0.55,
      subtotal: 11.55,
      requiresPrescription: true,
      canIncrease: false,
      canDecrease: true,
    },
  ];
  const subtotal = round2(items.reduce((sum, item) => sum + item.subtotal, 0));
  return {
    id: 'ord_1',
    code: 'DEMO-001',
    status: 'PENDING',
    clientId: '1712345678',
    clientName: 'María Pérez',
    phone: '+593991111111',
    prescriptionId: 'rx_1',
    mode: 'pickup',
    zone: { id: 'uio-norte', label: 'Quito — zona norte' },
    deliveryAddress: null,
    deliveryFee: 0,
    items,
    subtotal,
    total: subtotal,
    reservation: { expiresAt: '2026-10-08T15:14:05Z', secondsLeft: 540, active: true },
    fulfillments: [
      {
        pharmacyId: 'med-norte',
        pharmacyName: 'Medicity Demo Norte',
        address: 'Av. Ficticia A 123',
        status: 'NOTIFIED',
        eta: null,
        etaMinutes: 20,
        items: items.map((item) => ({ id: item.id, medicine: item.medicine, brand: item.brand, quantity: item.quantity })),
      },
    ],
    delivery: null,
    payment: null,
    createdAt: '2026-10-08T15:04:05Z',
    paidAt: null,
    ...overrides,
  };
}

export function round2(value: number): number {
  return Math.round(value * 100) / 100;
}

export function clone<T>(value: T): T {
  return structuredClone(value);
}
