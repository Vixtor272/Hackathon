import { ApiError, isPaid, type Order, type OrderItem, type PaymentOptions } from '../../domain';
import type { OrderGateway } from '../ports';

/**
 * Cart confirmation page logic. The backend enforces the quantity rules; this use case
 * mirrors them (`canIncrease` / `canDecrease`) so the UI fails fast with the same codes.
 */
export class Checkout {
  constructor(private readonly orders: OrderGateway) {}

  load(orderId: string): Promise<Order> {
    return this.orders.getOrder(orderId);
  }

  increase(order: Order, item: OrderItem): Promise<Order> {
    if (item.requiresPrescription) {
      return Promise.reject(new ApiError('RX_INCREASE_NOT_ALLOWED', 'Los medicamentos bajo receta no se pueden aumentar desde esta página'));
    }
    if (!item.canIncrease) {
      return Promise.reject(new ApiError('INSUFFICIENT_STOCK', 'No hay stock adicional para este producto'));
    }
    return this.changeQuantity(order, item.id, item.quantity + 1);
  }

  decrease(order: Order, item: OrderItem): Promise<Order> {
    if (!item.canDecrease) {
      return Promise.reject(new ApiError('INVALID_STATE', 'Este producto no se puede reducir'));
    }
    return this.changeQuantity(order, item.id, item.quantity - 1);
  }

  changeQuantity(order: Order, itemId: string, quantity: number): Promise<Order> {
    if (!Number.isInteger(quantity) || quantity < 0) {
      return Promise.reject(new ApiError('VALIDATION', 'La cantidad debe ser un entero mayor o igual a cero'));
    }
    if (isPaid(order)) {
      return Promise.reject(new ApiError('INVALID_STATE', 'El pedido ya fue pagado; el carrito no se puede modificar'));
    }
    if (!order.reservation.active) {
      return Promise.reject(new ApiError('RESERVATION_EXPIRED', 'La reserva venció. Renueva la reserva para seguir comprando'));
    }
    return this.orders.updateItemQuantity(order.id, itemId, quantity);
  }

  cancel(orderId: string): Promise<Order> {
    return this.orders.cancelOrder(orderId);
  }

  renew(orderId: string): Promise<Order> {
    return this.orders.renewOrder(orderId);
  }

  loadPaymentOptions(orderId: string): Promise<PaymentOptions> {
    return this.orders.getPaymentOptions(orderId);
  }

  /** True when the backend would accept a payment for this cart. */
  canPay(order: Order): boolean {
    return order.status === 'PENDING' && order.items.length > 0 && order.reservation.active;
  }
}
