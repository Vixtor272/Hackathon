import { describe, expect, it } from 'vitest';
import { InMemoryOrderGateway, sampleOrder } from '../../adapters/memory';
import { isApiError } from '../../domain';
import { Checkout } from './Checkout';

function setup() {
  const orders = new InMemoryOrderGateway([sampleOrder()]);
  return { orders, checkout: new Checkout(orders) };
}

describe('Checkout quantity rules', () => {
  it('refuses to raise a prescription item above the prescribed quantity without calling the backend', async () => {
    const { orders, checkout } = setup();
    const order = await checkout.load('ord_1');
    const rx = order.items.find((item) => item.requiresPrescription);
    if (!rx) throw new Error('fixture needs a prescription item');

    await expect(checkout.increase(order, rx)).rejects.toSatisfy(
      (error: unknown) => isApiError(error) && error.code === 'RX_INCREASE_NOT_ALLOWED',
    );
    const unchanged = await orders.getOrder('ord_1');
    expect(unchanged.items.find((item) => item.id === rx.id)?.quantity).toBe(rx.quantity);
  });

  it('lets a prescription item go down and back up to the prescribed quantity', async () => {
    const { checkout } = setup();
    let order = await checkout.load('ord_1');
    const rx = order.items.find((item) => item.requiresPrescription);
    if (!rx) throw new Error('fixture needs a prescription item');

    order = await checkout.decrease(order, rx);
    order = await checkout.decrease(order, order.items.find((item) => item.id === rx.id)!);
    let line = order.items.find((item) => item.id === rx.id)!;
    expect(line.quantity).toBe(rx.prescribedQuantity - 2);
    expect(line.canIncrease).toBe(true);

    order = await checkout.increase(order, line);
    order = await checkout.increase(order, order.items.find((item) => item.id === rx.id)!);
    line = order.items.find((item) => item.id === rx.id)!;
    expect(line.quantity).toBe(rx.prescribedQuantity);
    expect(line.canIncrease).toBe(false);
    await expect(checkout.increase(order, line)).rejects.toSatisfy(
      (error: unknown) => isApiError(error) && error.code === 'RX_INCREASE_NOT_ALLOWED',
    );
  });

  it('increases an over-the-counter item and recalculates totals', async () => {
    const { checkout } = setup();
    const order = await checkout.load('ord_1');
    const otc = order.items.find((item) => !item.requiresPrescription);
    if (!otc) throw new Error('fixture needs an over-the-counter item');

    const updated = await checkout.increase(order, otc);
    const item = updated.items.find((candidate) => candidate.id === otc.id);
    expect(item?.quantity).toBe(otc.quantity + 1);
    expect(item?.subtotal).toBeCloseTo((otc.quantity + 1) * otc.unitPrice, 2);
    expect(updated.total).toBeCloseTo(updated.items.reduce((sum, i) => sum + i.subtotal, 0), 2);
  });

  it('keeps a prescription item at zero and lets it grow back up to the prescribed quantity', async () => {
    const { checkout } = setup();
    let order = await checkout.load('ord_1');
    const rx = order.items.find((item) => item.requiresPrescription);
    if (!rx) throw new Error('fixture needs a prescription item');

    order = await checkout.changeQuantity(order, rx.id, 0);
    let line = order.items.find((item) => item.id === rx.id)!;
    expect(line).toMatchObject({ quantity: 0, subtotal: 0, canIncrease: true, canDecrease: false });

    for (let i = 0; i < rx.prescribedQuantity; i++) {
      order = await checkout.increase(order, order.items.find((item) => item.id === rx.id)!);
    }
    line = order.items.find((item) => item.id === rx.id)!;
    expect(line.quantity).toBe(rx.prescribedQuantity);
    expect(line.canIncrease).toBe(false);
    await expect(checkout.increase(order, line)).rejects.toSatisfy(
      (error: unknown) => isApiError(error) && error.code === 'RX_INCREASE_NOT_ALLOWED',
    );
  });

  it('cannot pay a cart with every item at zero', async () => {
    const { checkout } = setup();
    let order = await checkout.load('ord_1');
    for (const item of order.items) order = await checkout.changeQuantity(order, item.id, 0);
    expect(order.items.length).toBeGreaterThan(0);
    expect(order.total).toBe(0);
    expect(checkout.canPay(order)).toBe(false);
  });

  it('sets an item aside when its quantity reaches zero', async () => {
    const { checkout } = setup();
    let order = await checkout.load('ord_1');
    const otc = order.items.find((item) => !item.requiresPrescription);
    if (!otc) throw new Error('fixture needs an over-the-counter item');

    order = await checkout.changeQuantity(order, otc.id, 0);
    expect(order.items.find((item) => item.id === otc.id)?.quantity).toBe(0);
    expect(order.subtotal).toBeCloseTo(11.55, 2);
  });

  it('blocks edits once the reservation expired and allows renewing', async () => {
    const { orders, checkout } = setup();
    orders.mutate('ord_1', (order) => {
      order.status = 'EXPIRED';
      order.reservation = { ...order.reservation, active: false, secondsLeft: 0 };
    });
    const expired = await checkout.load('ord_1');
    expect(checkout.canPay(expired)).toBe(false);
    await expect(checkout.changeQuantity(expired, 'itm_1', 5)).rejects.toSatisfy(
      (error: unknown) => isApiError(error) && error.code === 'RESERVATION_EXPIRED',
    );

    const renewed = await checkout.renew('ord_1');
    expect(renewed.status).toBe('PENDING');
    expect(renewed.reservation.active).toBe(true);
    expect(checkout.canPay(renewed)).toBe(true);
  });
});
