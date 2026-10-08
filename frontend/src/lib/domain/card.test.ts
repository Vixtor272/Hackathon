import { describe, expect, it } from 'vitest';
import {
  detectBrand,
  formatCardNumber,
  formatExpiry,
  hasErrors,
  luhnValid,
  simulatedOutcome,
  TEST_CARDS,
  validateCard,
  type CardForm,
} from './card';

const NOW = new Date('2026-10-08T15:00:00Z');
const valid: CardForm = { number: '4242 4242 4242 4242', holder: 'María Pérez', expiry: '12/30', cvv: '123' };

describe('card form', () => {
  it('detects the brand from the first digits', () => {
    expect(detectBrand('4242')).toBe('visa');
    expect(detectBrand('5555 5555')).toBe('mastercard');
    expect(detectBrand('2221 00')).toBe('mastercard');
    expect(detectBrand('3782 822463')).toBe('amex');
    expect(detectBrand('3056 93')).toBe('diners');
    expect(detectBrand('9999')).toBe('unknown');
  });

  it('checks the Luhn digit', () => {
    expect(luhnValid('4242424242424242')).toBe(true);
    expect(luhnValid('4242424242424241')).toBe(false);
  });

  it('formats number and expiry while typing', () => {
    expect(formatCardNumber('4242424242424242')).toBe('4242 4242 4242 4242');
    expect(formatCardNumber('55555555555544449999')).toBe('5555 5555 5555 4444'); // Mastercard stops at 16
    expect(formatCardNumber('378282246310005')).toBe('3782 822463 10005');
    expect(formatExpiry('1')).toBe('1');
    expect(formatExpiry('3')).toBe('03');
    expect(formatExpiry('1230')).toBe('12/30');
  });

  it('accepts a complete card and every test card', () => {
    expect(validateCard(valid, NOW)).toEqual({});
    for (const card of TEST_CARDS) {
      expect(hasErrors(validateCard({ ...valid, number: card.number }, NOW))).toBe(false);
    }
  });

  it('reports each wrong field', () => {
    const errors = validateCard({ number: '4242 4242 4242 4241', holder: '', expiry: '09/26', cvv: '12' }, NOW);
    expect(Object.keys(errors).sort()).toEqual(['cvv', 'expiry', 'holder', 'number']);
    expect(errors.expiry).toContain('vencida');
    expect(validateCard({ ...valid, expiry: '13/30' }, NOW).expiry).toContain('mes');
    expect(validateCard({ ...valid, expiry: '10/26' }, NOW)).toEqual({});
  });

  it('asks for a 4-digit code on Amex', () => {
    const amex = { ...valid, number: '3782 822463 10005' };
    expect(validateCard(amex, NOW).cvv).toContain('4');
    expect(validateCard({ ...amex, cvv: '1234' }, NOW)).toEqual({});
  });

  it('decides the sandbox outcome by test number', () => {
    expect(simulatedOutcome('4242424242424242')).toBe('approved');
    expect(simulatedOutcome('4000 0000 0000 0002')).toBe('rejected');
  });
});
