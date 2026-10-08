import { writable, type Readable } from 'svelte/store';

export interface PersistedValue extends Readable<string | null> {
  set(value: string | null): void;
}

/** A string store mirrored to sessionStorage (best effort: private windows may block it). */
function persisted(key: string): PersistedValue {
  let initial: string | null = null;
  try {
    initial = sessionStorage.getItem(key);
  } catch {
    initial = null;
  }
  const store = writable<string | null>(initial);
  return {
    subscribe: store.subscribe,
    set(value: string | null): void {
      store.set(value);
      try {
        if (value) sessionStorage.setItem(key, value);
        else sessionStorage.removeItem(key);
      } catch {
        /* storage unavailable: keep in memory only */
      }
    },
  };
}

/** Last order and payment the person touched, so the nav can deep-link to them. */
export const lastOrderId = persisted('farmi.lastOrderId');
export const lastPaymentId = persisted('farmi.lastPaymentId');
export const activePhone = persisted('farmi.activePhone');
