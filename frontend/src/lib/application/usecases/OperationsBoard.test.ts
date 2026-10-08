import { describe, expect, it } from 'vitest';
import {
  InMemoryCatalogGateway,
  InMemoryFulfillmentGateway,
  InMemoryNotificationGateway,
  InMemoryOrderGateway,
  sampleOrder,
} from '../../adapters/memory';
import { OperationsBoard } from './OperationsBoard';

function setup() {
  const orders = new InMemoryOrderGateway([sampleOrder({ status: 'PAID', paidAt: '2026-10-08T15:10:00Z' })]);
  const board = new OperationsBoard(
    new InMemoryFulfillmentGateway(orders.store),
    new InMemoryCatalogGateway(),
    new InMemoryNotificationGateway([
      { id: 'ntf_1', channel: 'pharmacy', recipient: 'med-norte', title: 'Nuevo pedido DEMO-001', body: '…', orderId: 'ord_1', at: '2026-10-08T15:10:00Z' },
    ]),
  );
  return { orders, board };
}

describe('OperationsBoard', () => {
  it('lists every pharmacy across zones', async () => {
    const { board } = setup();
    const pharmacies = await board.listAllPharmacies();
    expect(pharmacies.map((pharmacy) => pharmacy.id)).toEqual(['med-norte', 'eco-norte', 'eco-centro']);
  });

  it('walks a pickup through preparation to hand-over', async () => {
    const { board } = setup();
    expect(board.nextPharmacyStep('NOTIFIED')).toEqual({ next: 'PREPARING', label: 'En preparación' });
    expect(board.nextPharmacyStep('PICKED_UP')).toBeNull();

    let order = await board.advancePharmacy('ord_1', 'med-norte', 'NOTIFIED');
    expect(order.fulfillments[0]?.status).toBe('PREPARING');
    order = await board.advancePharmacy('ord_1', 'med-norte', 'PREPARING');
    expect(order.fulfillments[0]?.status).toBe('READY');
    expect(order.status).toBe('READY');
    await expect(board.advancePharmacy('ord_1', 'med-norte', 'PICKED_UP')).rejects.toThrow();
  });

  it('shows only the orders for the selected pharmacy and filters notifications', async () => {
    const { board } = setup();
    expect(await board.ordersForPharmacy('med-norte')).toHaveLength(1);
    expect(await board.ordersForPharmacy('eco-centro')).toHaveLength(0);
    expect(await board.ordersForCourier()).toHaveLength(0);
    expect(await board.loadNotifications({ channel: 'pharmacy' })).toHaveLength(1);
    expect(await board.loadNotifications({ channel: 'courier' })).toHaveLength(0);
  });
});
