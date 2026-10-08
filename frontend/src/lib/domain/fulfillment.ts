export type FulfillmentStatus = 'NOTIFIED' | 'PREPARING' | 'READY' | 'PICKED_UP';
export type DeliveryStatus = 'PENDING' | 'CONSOLIDATING' | 'DISPATCHED' | 'DELIVERED';

export interface Courier {
  name: string;
  phone: string;
}

export interface FulfillmentItem {
  id: string;
  medicine: string;
  brand: string;
  quantity: number;
}

/** Preparation of one order at one pharmacy (pickup orders may have two). */
export interface Fulfillment {
  pharmacyId: string;
  pharmacyName: string;
  address: string;
  status: FulfillmentStatus;
  eta: string | null;
  etaMinutes: number;
  items: FulfillmentItem[];
}

export interface Delivery {
  status: DeliveryStatus;
  eta: string | null;
  etaMinutes: number;
  courier: Courier;
}

export const FULFILLMENT_STATUS_LABELS: Record<FulfillmentStatus, string> = {
  NOTIFIED: 'Notificada',
  PREPARING: 'En preparación',
  READY: 'Listo para recoger',
  PICKED_UP: 'Entregado al cliente',
};

export const DELIVERY_STATUS_LABELS: Record<DeliveryStatus, string> = {
  PENDING: 'Pendiente de pago',
  CONSOLIDATING: 'Consolidando pedido',
  DISPATCHED: 'En reparto',
  DELIVERED: 'Entregado',
};
