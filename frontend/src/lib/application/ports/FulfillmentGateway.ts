import type { DeliveryStatus, FulfillmentStatus, Order } from '../../domain';

export interface FulfillmentFilter {
  pharmacyId?: string;
  role?: 'courier';
}

/** Driven port: cashier and courier operations after payment. */
export interface FulfillmentGateway {
  listOrders(filter?: FulfillmentFilter): Promise<Order[]>;
  updatePharmacyStatus(orderId: string, pharmacyId: string, status: FulfillmentStatus): Promise<Order>;
  updateDeliveryStatus(orderId: string, status: DeliveryStatus): Promise<Order>;
}
