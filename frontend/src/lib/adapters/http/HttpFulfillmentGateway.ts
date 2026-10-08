import type { DeliveryStatus, FulfillmentStatus, Order } from '../../domain';
import type { FulfillmentFilter, FulfillmentGateway } from '../../application/ports';
import type { HttpClient } from './httpClient';

export class HttpFulfillmentGateway implements FulfillmentGateway {
  constructor(private readonly http: HttpClient) {}

  async listOrders(filter: FulfillmentFilter = {}): Promise<Order[]> {
    const { orders } = await this.http.get<{ orders: Order[] }>('/fulfillment/orders', {
      pharmacyId: filter.pharmacyId,
      role: filter.role,
    });
    return orders;
  }

  async updatePharmacyStatus(orderId: string, pharmacyId: string, status: FulfillmentStatus): Promise<Order> {
    const { order } = await this.http.post<{ order: Order }>(
      `/fulfillment/orders/${encodeURIComponent(orderId)}/pharmacies/${encodeURIComponent(pharmacyId)}/status`,
      { status },
    );
    return order;
  }

  async updateDeliveryStatus(orderId: string, status: DeliveryStatus): Promise<Order> {
    const { order } = await this.http.post<{ order: Order }>(
      `/fulfillment/orders/${encodeURIComponent(orderId)}/delivery/status`,
      { status },
    );
    return order;
  }
}
