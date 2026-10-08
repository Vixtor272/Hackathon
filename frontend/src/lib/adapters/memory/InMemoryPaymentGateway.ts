import { ApiError, type Bank, type Order, type Payment, type PaymentConfirmation, type PaymentMethod, type PaymentOutcome, type PaymentWithOrder } from '../../domain';
import type { PaymentGateway } from '../../application/ports';
import { clone } from './fixtures';

const BANKS: Bank[] = [
  { id: 'banco-demo-1', name: 'Banco Demo Pichincha' },
  { id: 'banco-demo-2', name: 'Banco Demo Guayaquil' },
];

/** Payment simulator with the backend's idempotency: approving twice deducts stock once. */
export class InMemoryPaymentGateway implements PaymentGateway {
  private readonly payments = new Map<string, Payment>();
  private counter = 0;
  /** How many times stock was deducted per order — must stay at 1. */
  readonly stockDeductions = new Map<string, number>();

  constructor(private readonly orders: Map<string, Order>) {}

  async createPayment(orderId: string, method: PaymentMethod): Promise<Payment> {
    const order = this.order(orderId);
    if (order.status !== 'PENDING') throw new ApiError('INVALID_STATE', 'El pedido no está pendiente de pago', 409);
    if (order.items.length === 0) throw new ApiError('EMPTY_CART', 'El carrito está vacío', 409);
    if (!order.reservation.active) throw new ApiError('RESERVATION_EXPIRED', 'La reserva venció', 409);

    for (const previous of this.payments.values()) {
      if (previous.orderId === orderId && previous.status === 'PENDING') previous.status = 'INVALIDATED';
    }
    this.counter += 1;
    const payment: Payment = {
      id: `pay_${this.counter}`,
      orderId,
      method,
      amount: order.total,
      status: 'PENDING',
      link: method === 'deuna' ? `http://localhost:5173/deuna/pay_${this.counter}` : null,
      bankId: null,
      createdAt: new Date(0).toISOString(),
      confirmedAt: null,
    };
    this.payments.set(payment.id, payment);
    order.payment = { id: payment.id, method, status: 'PENDING' };
    return clone(payment);
  }

  async getPayment(paymentId: string): Promise<PaymentWithOrder> {
    const payment = this.payment(paymentId);
    const order = this.order(payment.orderId);
    return { payment: clone(payment), order: { id: order.id, code: order.code, total: order.total, status: order.status } };
  }

  async listBanks(): Promise<Bank[]> {
    return clone(BANKS);
  }

  async confirmPayment(paymentId: string, outcome: PaymentOutcome, bankId?: string): Promise<PaymentConfirmation> {
    const payment = this.payment(paymentId);
    const order = this.order(payment.orderId);
    this.syncInvalidation(payment, order);
    if (payment.status === 'INVALIDATED') throw new ApiError('PAYMENT_INVALIDATED', 'El pago fue invalidado', 409);
    if (payment.status === 'PENDING') {
      payment.status = outcome === 'approved' ? 'APPROVED' : 'REJECTED';
      payment.bankId = bankId ?? null;
      payment.confirmedAt = new Date(0).toISOString();
      if (outcome === 'approved') {
        order.status = 'PAID';
        order.paidAt = payment.confirmedAt;
        order.reservation = { ...order.reservation, active: false, secondsLeft: 0 };
        this.stockDeductions.set(order.id, (this.stockDeductions.get(order.id) ?? 0) + 1);
      }
      order.payment = { id: payment.id, method: payment.method, status: payment.status };
    }
    return { payment: clone(payment), order: clone(order) };
  }

  /** Mirrors the backend: a cart change (new total) or the order gateway invalidates a pending intent. */
  private syncInvalidation(payment: Payment, order: Order): void {
    if (payment.status !== 'PENDING') return;
    const refInvalidated = order.payment?.id === payment.id && order.payment.status === 'INVALIDATED';
    if (refInvalidated || payment.amount !== order.total || !order.reservation.active) {
      payment.status = 'INVALIDATED';
    }
  }

  private payment(paymentId: string): Payment {
    const payment = this.payments.get(paymentId);
    if (!payment) throw new ApiError('NOT_FOUND', 'Pago no encontrado', 404);
    return payment;
  }

  private order(orderId: string): Order {
    const order = this.orders.get(orderId);
    if (!order) throw new ApiError('NOT_FOUND', 'Pedido no encontrado', 404);
    return order;
  }
}
