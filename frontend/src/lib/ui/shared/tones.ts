import type { DeliveryStatus, FulfillmentStatus, OrderStatus, PaymentStatus, SampleExpectation } from '../../domain';

export type Tone = 'neutral' | 'info' | 'success' | 'warning' | 'danger';

export function orderTone(status: OrderStatus): Tone {
  switch (status) {
    case 'PENDING':
      return 'warning';
    case 'PAID':
    case 'PREPARING':
    case 'DISPATCHED':
      return 'info';
    case 'READY':
    case 'DELIVERED':
      return 'success';
    case 'CANCELLED':
    case 'EXPIRED':
      return 'danger';
  }
}

export function fulfillmentTone(status: FulfillmentStatus): Tone {
  switch (status) {
    case 'NOTIFIED':
      return 'neutral';
    case 'PREPARING':
      return 'info';
    case 'READY':
    case 'PICKED_UP':
      return 'success';
  }
}

export function deliveryTone(status: DeliveryStatus): Tone {
  switch (status) {
    case 'PENDING':
      return 'warning';
    case 'CONSOLIDATING':
      return 'neutral';
    case 'DISPATCHED':
      return 'info';
    case 'DELIVERED':
      return 'success';
  }
}

export function paymentTone(status: PaymentStatus): Tone {
  switch (status) {
    case 'PENDING':
      return 'warning';
    case 'APPROVED':
      return 'success';
    case 'REJECTED':
    case 'INVALIDATED':
      return 'danger';
  }
}

export const EXPECTATION_LABELS: Record<SampleExpectation, { label: string; tone: Tone }> = {
  valid: { label: 'Válida', tone: 'success' },
  doctor_inactive: { label: 'Médico inactivo', tone: 'warning' },
  doctor_unknown: { label: 'Médico no registrado', tone: 'warning' },
  incomplete: { label: 'Incompleta', tone: 'warning' },
  illegible: { label: 'Ilegible', tone: 'danger' },
};
