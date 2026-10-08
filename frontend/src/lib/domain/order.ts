import type { Zone } from './catalog';
import type { Delivery, Fulfillment } from './fulfillment';
import type { PaymentMethod, PaymentStatus } from './payment';

export type OrderStatus =
  | 'PENDING'
  | 'PAID'
  | 'PREPARING'
  | 'READY'
  | 'DISPATCHED'
  | 'DELIVERED'
  | 'CANCELLED'
  | 'EXPIRED';

export type OrderMode = 'pickup' | 'delivery';

export interface OrderItem {
  id: string;
  sku: string;
  medicine: string;
  brand: string;
  presentation: string;
  unitLabel: string;
  pharmacyId: string;
  pharmacyName: string;
  quantity: number;
  prescribedQuantity: number;
  unitPrice: number;
  subtotal: number;
  requiresPrescription: boolean;
  canIncrease: boolean;
  canDecrease: boolean;
}

export interface Reservation {
  expiresAt: string;
  secondsLeft: number;
  active: boolean;
}

export interface OrderPaymentRef {
  id: string;
  method: PaymentMethod;
  status: PaymentStatus;
}

export interface Order {
  id: string;
  code: string;
  status: OrderStatus;
  clientId: string;
  clientName: string;
  phone: string;
  prescriptionId: string;
  mode: OrderMode;
  zone: Zone;
  deliveryAddress: string | null;
  deliveryFee: number;
  items: OrderItem[];
  subtotal: number;
  total: number;
  reservation: Reservation;
  fulfillments: Fulfillment[];
  delivery: Delivery | null;
  payment: OrderPaymentRef | null;
  createdAt: string;
  paidAt: string | null;
}

export interface PaymentOption {
  method: PaymentMethod;
  label: string;
  description: string;
}

export interface PaymentOptions {
  checkoutUrl: string;
  options: PaymentOption[];
}

export const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
  PENDING: 'Pendiente de pago',
  PAID: 'Pagado',
  PREPARING: 'En preparación',
  READY: 'Listo',
  DISPATCHED: 'En reparto',
  DELIVERED: 'Entregado',
  CANCELLED: 'Cancelado',
  EXPIRED: 'Reserva vencida',
};

/** Statuses from which the order has already been paid (stock deducted). */
export const PAID_STATUSES: ReadonlySet<OrderStatus> = new Set(['PAID', 'PREPARING', 'READY', 'DISPATCHED', 'DELIVERED']);

export function isPaid(order: Pick<Order, 'status'>): boolean {
  return PAID_STATUSES.has(order.status);
}

/** True when there is nothing to buy: lines set aside at zero do not count. */
export function isEmptyCart(order: Pick<Order, 'items'>): boolean {
  return order.items.every((item) => item.quantity === 0);
}

export function isClosed(order: Pick<Order, 'status'>): boolean {
  return order.status === 'CANCELLED' || order.status === 'EXPIRED';
}
