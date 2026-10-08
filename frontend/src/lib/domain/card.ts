import type { PaymentOutcome } from './payment';

/**
 * Card form rules of the payment simulator. Card data is validated in the
 * browser and never sent to the backend: the simulator only needs the outcome,
 * which is decided by the test card number (like a processor's sandbox).
 */

export type CardBrand = 'visa' | 'mastercard' | 'amex' | 'diners' | 'discover' | 'unknown';

export interface CardForm {
  number: string;
  holder: string;
  expiry: string; // "MM/AA"
  cvv: string;
}

export type CardField = keyof CardForm;
export type CardErrors = Partial<Record<CardField, string>>;

export const EMPTY_CARD: CardForm = { number: '', holder: '', expiry: '', cvv: '' };

export const CARD_BRAND_LABELS: Record<CardBrand, string> = {
  visa: 'Visa',
  mastercard: 'Mastercard',
  amex: 'American Express',
  diners: 'Diners Club',
  discover: 'Discover',
  unknown: 'Tarjeta',
};

export interface TestCard {
  number: string;
  outcome: PaymentOutcome;
  label: string;
}

/** Sandbox numbers shown on the page. Any other valid number is approved. */
export const TEST_CARDS: readonly TestCard[] = [
  { number: '4242 4242 4242 4242', outcome: 'approved', label: 'Visa · aprobada' },
  { number: '5555 5555 5555 4444', outcome: 'approved', label: 'Mastercard · aprobada' },
  { number: '3056 930902 5904', outcome: 'approved', label: 'Diners · aprobada' },
  { number: '4000 0000 0000 0002', outcome: 'rejected', label: 'Visa · fondos insuficientes' },
];

const DECLINED = new Set(TEST_CARDS.filter((card) => card.outcome === 'rejected').map((card) => digitsOf(card.number)));

export function digitsOf(value: string): string {
  return value.replace(/\D/g, '');
}

export function detectBrand(number: string): CardBrand {
  const d = digitsOf(number);
  if (/^4/.test(d)) return 'visa';
  if (/^(5[1-5]|2(2[2-9]|[3-6]\d|7[01]|720))/.test(d)) return 'mastercard';
  if (/^3[47]/.test(d)) return 'amex';
  if (/^3(0[0-5]|[689])/.test(d)) return 'diners';
  if (/^6(011|5)/.test(d)) return 'discover';
  return 'unknown';
}

function lengthsFor(brand: CardBrand): number[] {
  switch (brand) {
    case 'amex':
      return [15];
    case 'diners':
      return [14, 16];
    case 'visa':
      return [13, 16, 19];
    default:
      return [16];
  }
}

export function cvvLength(brand: CardBrand): number {
  return brand === 'amex' ? 4 : 3;
}

/** Luhn checksum, the check digit every real card number carries. */
export function luhnValid(number: string): boolean {
  const d = digitsOf(number);
  if (d.length < 12) return false;
  let sum = 0;
  for (let i = 0; i < d.length; i += 1) {
    let n = Number(d[d.length - 1 - i]);
    if (i % 2 === 1) {
      n *= 2;
      if (n > 9) n -= 9;
    }
    sum += n;
  }
  return sum % 10 === 0;
}

/** "4242424242424242" → "4242 4242 4242 4242" (Amex/Diners 4-6-5 / 4-6-4). */
export function formatCardNumber(raw: string): string {
  const brand = detectBrand(raw);
  const max = Math.max(...lengthsFor(brand));
  const d = digitsOf(raw).slice(0, max);
  const groups = brand === 'amex' || (brand === 'diners' && d.length <= 14) ? [4, 6, 5] : [4, 4, 4, 4, 3];
  const out: string[] = [];
  let at = 0;
  for (const size of groups) {
    if (at >= d.length) break;
    out.push(d.slice(at, at + size));
    at += size;
  }
  return out.join(' ');
}

/** "0128" → "01/28"; keeps what the person typed while they type. */
export function formatExpiry(raw: string): string {
  let d = digitsOf(raw).slice(0, 4);
  if (d.length === 1 && Number(d) > 1) d = `0${d}`;
  return d.length > 2 ? `${d.slice(0, 2)}/${d.slice(2)}` : d;
}

export function validateCard(form: CardForm, now: Date = new Date()): CardErrors {
  const errors: CardErrors = {};
  const digits = digitsOf(form.number);
  const brand = detectBrand(digits);

  if (digits.length === 0) errors.number = 'Ingresa el número de la tarjeta';
  else if (!lengthsFor(brand).includes(digits.length) || !luhnValid(digits)) errors.number = 'El número de tarjeta no es válido';

  const holder = form.holder.trim();
  if (holder.length === 0) errors.holder = 'Ingresa el nombre como aparece en la tarjeta';
  else if (!/^[\p{L} .'-]{3,}$/u.test(holder)) errors.holder = 'Usa solo letras y espacios';

  const match = /^(\d{2})\/(\d{2})$/.exec(form.expiry.trim());
  if (!match) {
    errors.expiry = 'Usa el formato MM/AA';
  } else {
    const month = Number(match[1]);
    const year = 2000 + Number(match[2]);
    if (month < 1 || month > 12) errors.expiry = 'El mes debe estar entre 01 y 12';
    else if (year * 12 + month < now.getFullYear() * 12 + now.getMonth() + 1) errors.expiry = 'La tarjeta está vencida';
    else if (year > now.getFullYear() + 20) errors.expiry = 'La fecha de vencimiento no es válida';
  }

  const cvv = digitsOf(form.cvv);
  if (cvv.length !== cvvLength(brand) || cvv.length !== form.cvv.trim().length) {
    errors.cvv = `El código de seguridad tiene ${cvvLength(brand)} dígitos`;
  }
  return errors;
}

export function hasErrors(errors: CardErrors): boolean {
  return Object.keys(errors).length > 0;
}

/** Sandbox decision: the declined test numbers are rejected, everything else approved. */
export function simulatedOutcome(number: string): PaymentOutcome {
  return DECLINED.has(digitsOf(number)) ? 'rejected' : 'approved';
}

/** "•••• 4242" for receipts and confirmations. */
export function maskedNumber(number: string): string {
  const d = digitsOf(number);
  return d.length >= 4 ? `•••• ${d.slice(-4)}` : '';
}
