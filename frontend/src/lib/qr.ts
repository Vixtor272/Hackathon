import qrcode from 'qrcode-generator';

/**
 * Encodes text as a QR matrix (true = dark module). The library stays behind
 * this function so the UI only deals with rows of booleans.
 */
export function qrMatrix(text: string): boolean[][] {
  const qr = qrcode(0, 'M');
  qr.addData(text, 'Byte');
  qr.make();
  const size = qr.getModuleCount();
  return Array.from({ length: size }, (_, row) => Array.from({ length: size }, (_, col) => qr.isDark(row, col)));
}
