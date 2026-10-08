import {
  FetchHttpClient,
  HttpCatalogGateway,
  HttpChatGateway,
  HttpFulfillmentGateway,
  HttpNotificationGateway,
  HttpOrderGateway,
  HttpPaymentGateway,
  type HttpClient,
} from './adapters/http';
import type {
  CatalogGateway,
  ChatGateway,
  FulfillmentGateway,
  NotificationGateway,
  OrderGateway,
  PaymentGateway,
} from './application/ports';
import { ChatSession, Checkout, OperationsBoard, PaymentFlow } from './application/usecases';

export interface Gateways {
  chat: ChatGateway;
  orders: OrderGateway;
  payments: PaymentGateway;
  fulfillment: FulfillmentGateway;
  catalog: CatalogGateway;
  notifications: NotificationGateway;
}

export interface Container {
  gateways: Gateways;
  chatSession: ChatSession;
  checkout: Checkout;
  paymentFlow: PaymentFlow;
  operations: OperationsBoard;
}

/** Composition root: wires driven adapters into the use cases. The UI only imports `app`. */
export function createContainer(gateways: Gateways): Container {
  return {
    gateways,
    chatSession: new ChatSession(gateways.chat),
    checkout: new Checkout(gateways.orders),
    paymentFlow: new PaymentFlow(gateways.payments),
    operations: new OperationsBoard(gateways.fulfillment, gateways.catalog, gateways.notifications),
  };
}

export function createHttpGateways(http: HttpClient = new FetchHttpClient('/api/v1')): Gateways {
  return {
    chat: new HttpChatGateway(http),
    orders: new HttpOrderGateway(http),
    payments: new HttpPaymentGateway(http),
    fulfillment: new HttpFulfillmentGateway(http),
    catalog: new HttpCatalogGateway(http),
    notifications: new HttpNotificationGateway(http),
  };
}

export const app: Container = createContainer(createHttpGateways());
