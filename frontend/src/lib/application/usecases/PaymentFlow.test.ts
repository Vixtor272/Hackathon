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

  it('pays with the card form: test numbers decide the outcome, invalid data never reaches the gateway', async () => {
    const { flow } = setup();
    const card = { number: '4000 0000 0000 0002', holder: 'María Pérez', expiry: '12/30', cvv: '123' };
    const now = new Date('2026-10-08T15:00:00Z');

    await expect(flow.payWithCardDetails('ord_1', { ...card, cvv: '1' }, now)).rejects.toSatisfy(
      (error: unknown) => isApiError(error) && error.code === 'VALIDATION',
    );
    const declined = await flow.payWithCardDetails('ord_1', card, now);
    expect(declined.payment.status).toBe('REJECTED');
    const approved = await flow.payWithCardDetails('ord_1', { ...card, number: '4242 4242 4242 4242' }, now);
    expect(approved.order.status).toBe('PAID');
  });

  it('offers DeUna as a QR with a random reference and pays it without choosing a bank', async () => {
    const orders = new InMemoryOrderGateway([sampleOrder()]);
    const flow = new PaymentFlow(new InMemoryPaymentGateway(orders.store), () => 'DU-TESTREF1');
    const qr = await flow.startDeUnaQr('ord_1');
    expect(qr.reference).toBe('DU-TESTREF1');
    expect(qr.qrPayload).toContain(`/deuna/${qr.payment.id}`);
    expect(qr.qrPayload).toContain('ref=DU-TESTREF1');

    const paid = await flow.confirmDeUnaQr(qr.payment.id, 'approved');
    expect(paid.payment.status).toBe('APPROVED');
    expect(paid.order.status).toBe('PAID');
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
