/** WhatsApp text markup, as Farmi writes it: `*bold*` spans and a title on the first line. */
export interface MarkupSpan {
  text: string;
  bold: boolean;
}

export interface MarkupLine {
  spans: MarkupSpan[];
  /** First line of the message when it carries a bold heading ("💊 *Paracetamol*"). */
  title: boolean;
}

const BOLD = /\*(\S(?:[^*\n]*\S)?)\*/g;

function parseSpans(line: string): MarkupSpan[] {
  const spans: MarkupSpan[] = [];
  let last = 0;
  for (const match of line.matchAll(BOLD)) {
    if (match.index > last) spans.push({ text: line.slice(last, match.index), bold: false });
    spans.push({ text: match[1] ?? '', bold: true });
    last = match.index + match[0].length;
  }
  if (last < line.length) spans.push({ text: line.slice(last), bold: false });
  return spans;
}

/** Splits a message into lines of plain / bold spans; unmatched asterisks stay as typed. */
export function parseMarkup(text: string): MarkupLine[] {
  return text.split('\n').map((line, index) => {
    const spans = parseSpans(line);
    return { spans, title: index === 0 && spans.some((span) => span.bold) };
  });
}
