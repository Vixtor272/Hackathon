import type {
  DeliveryStatus,
  FulfillmentStatus,
  Notification,
  Order,
  Pharmacy,
} from '../../domain';
import type { CatalogGateway, FulfillmentGateway, NotificationFilter, NotificationGateway } from '../ports';

export interface Step<S> {
  next: S;
  label: string;
}

const PHARMACY_STEPS: Partial<Record<FulfillmentStatus, Step<FulfillmentStatus>>> = {
  NOTIFIED: { next: 'PREPARING', label: 'En preparación' },
  PREPARING: { next: 'READY', label: 'Listo para recoger' },
  READY: { next: 'PICKED_UP', label: 'Entregado al cliente' },
};

const DELIVERY_STEPS: Partial<Record<DeliveryStatus, Step<DeliveryStatus>>> = {
  CONSOLIDATING: { next: 'DISPATCHED', label: 'En reparto' },
  DISPATCHED: { next: 'DELIVERED', label: 'Entregado' },
};

/** Cashier / courier board: lists paid orders and advances their preparation. */
export class OperationsBoard {
  constructor(
    private readonly fulfillment: FulfillmentGateway,
    private readonly catalog: CatalogGateway,
    private readonly notifications: NotificationGateway,
  ) {}

  async listAllPharmacies(): Promise<Pharmacy[]> {
    const zones = await this.catalog.listZones();
    const perZone = await Promise.all(zones.map((zone) => this.catalog.listPharmacies(zone.id)));
    return perZone.flat();
  }

  ordersForPharmacy(pharmacyId: string): Promise<Order[]> {
    return this.fulfillment.listOrders({ pharmacyId });
  }

  ordersForCourier(): Promise<Order[]> {
    return this.fulfillment.listOrders({ role: 'courier' });
  }

  nextPharmacyStep(status: FulfillmentStatus): Step<FulfillmentStatus> | null {
    return PHARMACY_STEPS[status] ?? null;
  }

  nextDeliveryStep(status: DeliveryStatus): Step<DeliveryStatus> | null {
    return DELIVERY_STEPS[status] ?? null;
  }

  advancePharmacy(orderId: string, pharmacyId: string, current: FulfillmentStatus): Promise<Order> {
    const step = this.nextPharmacyStep(current);
    if (!step) {
      return Promise.reject(new Error('Este retiro ya está completado'));
    }
    return this.fulfillment.updatePharmacyStatus(orderId, pharmacyId, step.next);
  }

  advanceDelivery(orderId: string, current: DeliveryStatus): Promise<Order> {
    const step = this.nextDeliveryStep(current);
    if (!step) {
      return Promise.reject(new Error('Esta entrega ya está completada'));
    }
    return this.fulfillment.updateDeliveryStatus(orderId, step.next);
  }

  loadNotifications(filter?: NotificationFilter): Promise<Notification[]> {
    return this.notifications.listNotifications(filter);
  }
}
