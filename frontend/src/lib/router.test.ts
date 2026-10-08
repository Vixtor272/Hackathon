import { describe, expect, it } from 'vitest';
import { matchRoute } from './router';

describe('matchRoute', () => {
  it('maps the four application paths', () => {
    expect(matchRoute('/')).toEqual({ name: 'whatsapp' });
    expect(matchRoute('/checkout/ord_1')).toEqual({ name: 'checkout', orderId: 'ord_1' });
    expect(matchRoute('/deuna/pay_7/')).toEqual({ name: 'deuna', paymentId: 'pay_7' });
    expect(matchRoute('/operaciones')).toEqual({ name: 'operations' });
  });

  it('falls back to notFound', () => {
    expect(matchRoute('/checkout')).toEqual({ name: 'notFound', path: '/checkout' });
    expect(matchRoute('/otra/cosa')).toEqual({ name: 'notFound', path: '/otra/cosa' });
  });
});
