const money = new Intl.NumberFormat('es-EC', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 });
const time = new Intl.DateTimeFormat('es-EC', { hour: '2-digit', minute: '2-digit' });
const dateTime = new Intl.DateTimeFormat('es-EC', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' });

export function formatMoney(value: number): string {
  return money.format(value);
}

export function formatTime(iso: string | null | undefined): string {
  if (!iso) return '';
  const date = new Date(iso);
  return Number.isNaN(date.getTime()) ? '' : time.format(date);
}

export function formatDateTime(iso: string | null | undefined): string {
  if (!iso) return '—';
  const date = new Date(iso);
  return Number.isNaN(date.getTime()) ? '—' : dateTime.format(date);
}

/** "09:41" style countdown from a number of seconds. */
export function formatCountdown(totalSeconds: number): string {
  const seconds = Math.max(0, Math.floor(totalSeconds));
  const minutes = Math.floor(seconds / 60);
  const rest = seconds % 60;
  return `${String(minutes).padStart(2, '0')}:${String(rest).padStart(2, '0')}`;
}

/** Minutes until an ISO timestamp, never negative; null when there is no timestamp. */
export function minutesUntil(iso: string | null | undefined, now: Date = new Date()): number | null {
  if (!iso) return null;
  const target = new Date(iso).getTime();
  if (Number.isNaN(target)) return null;
  return Math.max(0, Math.round((target - now.getTime()) / 60_000));
}
