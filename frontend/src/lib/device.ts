/** Where the checkout page is open: it decides how DeUna is offered. */
export type DeviceKind = 'desktop' | 'mobile';

export interface DeviceSignals {
  userAgent: string;
  /** `navigator.userAgentData.mobile` where the browser exposes it. */
  uaMobile?: boolean;
  maxTouchPoints: number;
  /** `(pointer: fine)` media query: a mouse or trackpad drives the page. */
  finePointer: boolean;
}

const MOBILE_UA = /Android|iPhone|iPod|iPad|Mobile|webOS|BlackBerry|IEMobile|Opera Mini/i;

/**
 * A computer has a fine pointer and no mobile user agent. iPadOS pretends to
 * be a Mac, so a "Macintosh" with a touch screen counts as mobile.
 */
export function classifyDevice(signals: DeviceSignals): DeviceKind {
  if (signals.uaMobile === true) return 'mobile';
  if (MOBILE_UA.test(signals.userAgent)) return 'mobile';
  if (/Macintosh/i.test(signals.userAgent) && signals.maxTouchPoints > 1) return 'mobile';
  if (!signals.finePointer && signals.maxTouchPoints > 0) return 'mobile';
  return 'desktop';
}

/** Reads the signals from the running browser (desktop when they are unavailable). */
export function detectDevice(): DeviceKind {
  if (typeof navigator === 'undefined') return 'desktop';
  const uaData = (navigator as Navigator & { userAgentData?: { mobile?: boolean } }).userAgentData;
  return classifyDevice({
    userAgent: navigator.userAgent,
    uaMobile: uaData?.mobile,
    maxTouchPoints: navigator.maxTouchPoints ?? 0,
    finePointer: typeof matchMedia === 'function' ? matchMedia('(pointer: fine)').matches : true,
  });
}
