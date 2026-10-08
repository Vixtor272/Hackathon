import { ApiError, isPaid, type DeliveryStatus, type FulfillmentStatus, type Order } from '../../domain';
import type { FulfillmentFilter, FulfillmentGateway } from '../../application/ports';
import { clone } from './fixtures';

export class InMemoryFulfillmentGateway implements FulfillmentGateway {
  constructor(private readonly orders: Map<string, Order>) {}

  async listOrders(filter: FulfillmentFilter = {}): Promise<Order[]> {
    const paid = [...this.orders.values()].filter(isPaid);
    const selected = paid.filter((order) => {
      if (filter.role === 'courier') return order.mode === 'delivery';
      if (filter.pharmacyId) return order.fulfillments.some((f) => f.pharmacyId === filter.pharmacyId);
      return true;
    });
    return clone(selected);
  }

  async updatePharmacyStatus(orderId: string, pharmacyId: string, status: FulfillmentStatus): Promise<Order> {
    const order = this.find(orderId);
    const fulfillment = order.fulfillments.find((f) => f.pharmacyId === pharmacyId);
    if (!fulfillment) throw new ApiError('NOT_FOUND', 'La farmacia no participa en este pedido', 404);
    fulfillment.status = status;
    if (status === 'READY') fulfillment.eta = new Date(0).toISOString();
    order.status = status === 'PICKED_UP' ? 'DELIVERED' : status === 'READY' ? 'READY' : 'PREPARING';
    return clone(order);
  }

  async updateDeliveryStatus(orderId: string, status: DeliveryStatus): Promise<Order> {
    const order = this.find(orderId);
    if (!order.delivery) throw new ApiError('INVALID_STATE', 'El pedido no es a domicilio', 409);
    order.delivery.status = status;
    order.status = status === 'DELIVERED' ? 'DELIVERED' : 'DISPATCHED';
    return clone(order);
  }

  private find(orderId: string): Order {
    const order = this.orders.get(orderId);
    if (!order) throw new ApiError('NOT_FOUND', 'Pedido no encontrado', 404);
    if (!isPaid(order)) throw new ApiError('INVALID_STATE', 'El pedido aún no está pagado', 409);
    return order;
  }
}
