import { describe, expect, it } from 'vitest';
import { InMemoryOrderGateway, InMemoryPaymentGateway, sampleOrder } from '../../adapters/memory';
import { isApiError } from '../../domain';
import { Checkout } from './Checkout';
import { PaymentFlow } from './PaymentFlow';

function setup() {
  const orders = new InMemoryOrderGateway([sampleOrder()]);
  const payments = new InMemoryPaymentGateway(orders.store);
  return { orders, payments, flow: new PaymentFlow(payments), checkout: new Checkout(orders) };
}

describe('PaymentFlow', () => {
  it('approves a card payment and confirming again is idempotent', async () => {
    const { payments, flow } = setup();
    const first = await flow.payWithCard('ord_1', 'approved');
    expect(first.payment.status).toBe('APPROVED');
    expect(first.order.status).toBe('PAID');

    const again = await flow.confirm(first.payment.id, 'approved');
    expect(again.payment.id).toBe(first.payment.id);
    expect(again.order.status).toBe('PAID');
    expect(payments.stockDeductions.get('ord_1')).toBe(1);
  });

  it('keeps the order pending after a rejected payment so the client can retry', async () => {
    const { flow } = setup();
    const rejected = await flow.payWithCard('ord_1', 'rejected');
    expect(rejected.payment.status).toBe('REJECTED');
    expect(rejected.order.status).toBe('PENDING');

    const retry = await flow.payWithCard('ord_1', 'approved');
    expect(retry.order.status).toBe('PAID');
  });

  it('starts DeUna with a link and requires a bank to approve', async () => {
    const { flow } = setup();
    const { payment, link } = await flow.startDeUna('ord_1');
    expect(link).toContain(`/deuna/${payment.id}`);
    await expect(flow.confirmDeUna(payment.id, 'approved', null)).rejects.toSatisfy(
      (error: unknown) => isApiError(error) && error.code === 'VALIDATION',
    );
    const confirmed = await flow.confirmDeUna(payment.id, 'approved', 'banco-demo-1');
    expect(confirmed.payment.bankId).toBe('banco-demo-1');
    expect(confirmed.order.status).toBe('PAID');
  });

  it('invalidates a pending payment when the cart changes', async () => {
    const { flow, checkout } = setup();
    const { payment } = await flow.startDeUna('ord_1');
    const order = await checkout.load('ord_1');
    const otc = order.items.find((item) => !item.requiresPrescription);
    if (!otc) throw new Error('fixture needs an over-the-counter item');
    await checkout.decrease(order, otc);

    await expect(flow.confirm(payment.id, 'approved', 'banco-demo-1')).rejects.toSatisfy(
      (error: unknown) => isApiError(error) && error.code === 'PAYMENT_INVALIDATED',
    );
  });
});
