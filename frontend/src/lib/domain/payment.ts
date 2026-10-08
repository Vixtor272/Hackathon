import type { Order, OrderStatus } from './order';

export type PaymentMethod = 'card' | 'deuna';
export type PaymentStatus = 'PENDING' | 'APPROVED' | 'REJECTED' | 'INVALIDATED';
export type PaymentOutcome = 'approved' | 'rejected';

export interface Payment {
  id: string;
  orderId: string;
  method: PaymentMethod;
  amount: number;
  status: PaymentStatus;
  link: string | null;
  bankId: string | null;
  createdAt: string;
  confirmedAt: string | null;
}

export interface Bank {
  id: string;
  name: string;
}

export interface PaymentOrderSummary {
  id: string;
  code: string;
  total: number;
  status: OrderStatus;
}

export interface PaymentWithOrder {
  payment: Payment;
  order: PaymentOrderSummary;
}

export interface PaymentConfirmation {
  payment: Payment;
  order: Order;
}

export const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  card: 'Tarjeta de débito o crédito',
  deuna: 'DeUna',
};

export const PAYMENT_STATUS_LABELS: Record<PaymentStatus, string> = {
  PENDING: 'Pendiente',
  APPROVED: 'Aprobado',
  REJECTED: 'Rechazado',
  INVALIDATED: 'Invalidado',
};

/**
 * A short random payment reference such as "DU-7F3K9Q2A" for the DeUna QR
 * (no 0/O/1/I so it can be read aloud or typed without mistakes).
 */
export function randomReference(prefix = 'DU', length = 8, random: () => number = secureRandom): string {
  const alphabet = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
  let out = '';
  for (let i = 0; i < length; i += 1) out += alphabet[Math.floor(random() * alphabet.length)];
  return `${prefix}-${out}`;
}

function secureRandom(): number {
  const buf = new Uint32Array(1);
  crypto.getRandomValues(buf);
  return (buf[0] ?? 0) / 2 ** 32;
}
