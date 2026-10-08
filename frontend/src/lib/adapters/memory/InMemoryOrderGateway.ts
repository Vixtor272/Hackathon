import { ApiError, type Order, type PaymentOptions } from '../../domain';
import type { OrderGateway } from '../../application/ports';
import { clone, round2, sampleOrder } from './fixtures';

/** In-memory order store that applies the same cart rules as the backend. */
export class InMemoryOrderGateway implements OrderGateway {
  /** Shared with the other in-memory fakes so payment and fulfillment see the same orders. */
  readonly store = new Map<string, Order>();
  /** Extra stock available per SKU for `+` on over-the-counter items. */
  readonly extraStock = new Map<string, number>();

  constructor(orders: Order[] = [sampleOrder()]) {
    for (const order of orders) this.store.set(order.id, clone(order));
  }

  async getOrder(orderId: string): Promise<Order> {
    return clone(this.find(orderId));
  }

  async updateItemQuantity(orderId: string, itemId: string, quantity: number): Promise<Order> {
    const order = this.find(orderId);
    if (order.status !== 'PENDING') throw new ApiError('INVALID_STATE', 'El pedido no se puede modificar', 409);
    if (!order.reservation.active) throw new ApiError('RESERVATION_EXPIRED', 'La reserva venció', 409);
    const item = order.items.find((candidate) => candidate.id === itemId);
    if (!item) throw new ApiError('NOT_FOUND', 'Producto no encontrado', 404);

    if (quantity > item.quantity) {
      if (item.requiresPrescription) throw new ApiError('RX_INCREASE_NOT_ALLOWED', 'No se puede aumentar un medicamento bajo receta', 409);
      const extra = this.extraStock.get(item.sku) ?? Number.POSITIVE_INFINITY;
      if (quantity - item.quantity > extra) throw new ApiError('INSUFFICIENT_STOCK', 'No hay stock adicional', 409);
      if (Number.isFinite(extra)) this.extraStock.set(item.sku, extra - (quantity - item.quantity));
    }

    if (quantity === 0) {
      order.items = order.items.filter((candidate) => candidate.id !== itemId);
    } else {
      item.quantity = quantity;
      item.subtotal = round2(quantity * item.unitPrice);
      item.canIncrease = !item.requiresPrescription && (this.extraStock.get(item.sku) ?? 1) > 0;
    }
    this.recalculate(order);
    if (order.payment && order.payment.status === 'PENDING') order.payment = { ...order.payment, status: 'INVALIDATED' };
    return clone(order);
  }

  async cancelOrder(orderId: string): Promise<Order> {
    const order = this.find(orderId);
    order.status = 'CANCELLED';
    order.reservation = { ...order.reservation, active: false, secondsLeft: 0 };
    return clone(order);
  }

  async renewOrder(orderId: string): Promise<Order> {
    const order = this.find(orderId);
    if (order.status !== 'EXPIRED') throw new ApiError('INVALID_STATE', 'Solo se renuevan reservas vencidas', 409);
    order.status = 'PENDING';
    order.reservation = { expiresAt: '2026-10-08T15:24:05Z', secondsLeft: 600, active: true };
    return clone(order);
  }

  async getPaymentOptions(orderId: string): Promise<PaymentOptions> {
    this.find(orderId);
    return {
      checkoutUrl: `http://localhost:5173/checkout/${orderId}`,
      options: [
        { method: 'card', label: 'Tarjeta de débito o crédito', description: 'Simulador sin datos bancarios reales' },
        { method: 'deuna', label: 'DeUna', description: 'Paga con tu banco (simulado)' },
      ],
    };
  }

  /** Test helper: mutate a stored order (e.g. mark it paid or expired). */
  mutate(orderId: string, change: (order: Order) => void): void {
    change(this.find(orderId));
  }

  private find(orderId: string): Order {
    const order = this.store.get(orderId);
    if (!order) throw new ApiError('NOT_FOUND', 'Pedido no encontrado', 404);
    return order;
  }

  private recalculate(order: Order): void {
    order.subtotal = round2(order.items.reduce((sum, item) => sum + item.subtotal, 0));
    order.total = round2(order.subtotal + (order.mode === 'delivery' ? order.deliveryFee : 0));
  }
}
