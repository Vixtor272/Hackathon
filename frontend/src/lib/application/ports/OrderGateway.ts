import type { Order, PaymentOptions } from '../../domain';

/** Driven port: orders, cart changes and reservations. */
export interface OrderGateway {
  getOrder(orderId: string): Promise<Order>;
  updateItemQuantity(orderId: string, itemId: string, quantity: number): Promise<Order>;
  cancelOrder(orderId: string): Promise<Order>;
  renewOrder(orderId: string): Promise<Order>;
  getPaymentOptions(orderId: string): Promise<PaymentOptions>;
}
