import type { Order, PaymentOptions } from '../../domain';
import type { OrderGateway } from '../../application/ports';
import type { HttpClient } from './httpClient';

interface OrderEnvelope {
  order: Order;
}

export class HttpOrderGateway implements OrderGateway {
  constructor(private readonly http: HttpClient) {}

  async getOrder(orderId: string): Promise<Order> {
    const { order } = await this.http.get<OrderEnvelope>(`/orders/${encodeURIComponent(orderId)}`);
    return order;
  }

  async updateItemQuantity(orderId: string, itemId: string, quantity: number): Promise<Order> {
    const { order } = await this.http.patch<OrderEnvelope>(
      `/orders/${encodeURIComponent(orderId)}/items/${encodeURIComponent(itemId)}`,
      { quantity },
    );
    return order;
  }

  async cancelOrder(orderId: string): Promise<Order> {
    const { order } = await this.http.post<OrderEnvelope>(`/orders/${encodeURIComponent(orderId)}/cancel`);
    return order;
  }

  async renewOrder(orderId: string): Promise<Order> {
    const { order } = await this.http.post<OrderEnvelope>(`/orders/${encodeURIComponent(orderId)}/renew`);
    return order;
  }

  getPaymentOptions(orderId: string): Promise<PaymentOptions> {
    return this.http.get<PaymentOptions>(`/orders/${encodeURIComponent(orderId)}/payment-options`);
  }
}
