import { describe, expect, it } from 'vitest';
import { randomReference } from './domain';
import { qrMatrix } from './qr';

describe('qrMatrix', () => {
  it('builds a square matrix with the three finder patterns', () => {
    const m = qrMatrix('http://localhost:5173/deuna/pay_1?ref=DU-ABCDEFGH');
    const size = m.length;
    expect(size).toBeGreaterThanOrEqual(21);
    expect(m.every((row) => row.length === size)).toBe(true);
    for (const [r, c] of [[0, 0], [0, size - 7], [size - 7, 0]] as const) {
      expect(m[r]?.[c]).toBe(true); // outer corner of each finder square
      expect(m[r + 1]?.[c + 1]).toBe(false); // white ring
      expect(m[r + 3]?.[c + 3]).toBe(true); // dark centre
    }
  });
});

describe('randomReference', () => {
  it('uses the prefix and an unambiguous alphabet', () => {
    expect(randomReference('DU', 8, () => 0)).toBe('DU-AAAAAAAA');
    expect(randomReference()).toMatch(/^DU-[A-HJ-NP-Z2-9]{8}$/);
    expect(randomReference()).not.toBe(randomReference());
  });
});
