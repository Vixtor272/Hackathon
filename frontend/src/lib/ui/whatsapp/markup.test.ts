import { describe, expect, it } from 'vitest';
import { parseMarkup } from './markup';

describe('parseMarkup', () => {
  it('marks the bold spans and the title line', () => {
    const lines = parseMarkup('💊 *Paracetamol 500 mg*\nNecesitas *20 tabletas* hoy.');
    expect(lines[0]).toEqual({
      title: true,
      spans: [
        { text: '💊 ', bold: false },
        { text: 'Paracetamol 500 mg', bold: true },
      ],
    });
    expect(lines[1]).toEqual({
      title: false,
      spans: [
        { text: 'Necesitas ', bold: false },
        { text: '20 tabletas', bold: true },
        { text: ' hoy.', bold: false },
      ],
    });
  });

  it('keeps plain text, empty lines and stray asterisks as typed', () => {
    expect(parseMarkup('hola')).toEqual([{ title: false, spans: [{ text: 'hola', bold: false }] }]);
    expect(parseMarkup('a\n\nb')[1]).toEqual({ title: false, spans: [] });
    expect(parseMarkup('2 * 3 = 6')[0]?.spans).toEqual([{ text: '2 * 3 = 6', bold: false }]);
  });
});
