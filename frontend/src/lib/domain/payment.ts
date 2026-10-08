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
