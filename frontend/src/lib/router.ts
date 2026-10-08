import { writable, type Readable } from 'svelte/store';

/**
 * Routes of the customer app. The company back office (cashier / courier
 * board) is a separate app on its own port, see empresa/ and EmpresaApp.svelte.
 */
export type Route =
  | { name: 'whatsapp' }
  | { name: 'checkout'; orderId: string }
  | { name: 'deuna'; paymentId: string }
  | { name: 'notFound'; path: string };

/** Pure path → route mapping (tested in isolation). */
export function matchRoute(pathname: string): Route {
  const segments = pathname.split('/').filter((segment) => segment.length > 0);
  const [head, tail, ...rest] = segments;

  if (segments.length === 0) return { name: 'whatsapp' };
  if (head === 'checkout' && tail && rest.length === 0) return { name: 'checkout', orderId: decodeURIComponent(tail) };
  if (head === 'deuna' && tail && rest.length === 0) return { name: 'deuna', paymentId: decodeURIComponent(tail) };
  return { name: 'notFound', path: pathname };
}

export const ROUTES = {
  whatsapp: '/',
  checkout: (orderId: string) => `/checkout/${encodeURIComponent(orderId)}`,
  deuna: (paymentId: string) => `/deuna/${encodeURIComponent(paymentId)}`,
} as const;

const current = writable<Route>({ name: 'whatsapp' });

/** Current route as a readable store; components use `$route`. */
export const route: Readable<Route> = { subscribe: current.subscribe };

export function navigate(path: string, options: { replace?: boolean } = {}): void {
  if (options.replace) {
    history.replaceState(null, '', path);
  } else {
    history.pushState(null, '', path);
  }
  current.set(matchRoute(location.pathname));
}

/**
 * Opens an href: same-origin links stay inside the SPA (history API),
 * anything else is a full navigation.
 */
export function openLink(href: string): void {
  let url: URL;
  try {
    url = new URL(href, location.href);
  } catch {
    return;
  }
  if (url.origin === location.origin) {
    navigate(`${url.pathname}${url.search}${url.hash}`);
  } else {
    location.assign(url.href);
  }
}

/** Starts listening to history and intercepts same-origin anchor clicks. Returns a stop function. */
export function startRouter(): () => void {
  current.set(matchRoute(location.pathname));

  const onPopState = (): void => current.set(matchRoute(location.pathname));
  const onClick = (event: MouseEvent): void => {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    const anchor = (event.target as Element | null)?.closest('a[href]');
    if (!(anchor instanceof HTMLAnchorElement)) return;
    if (anchor.target && anchor.target !== '_self') return;
    if (anchor.hasAttribute('download') || anchor.getAttribute('rel') === 'external') return;
    const url = new URL(anchor.href, location.href);
    if (url.origin !== location.origin) return;
    event.preventDefault();
    navigate(`${url.pathname}${url.search}${url.hash}`);
  };

  window.addEventListener('popstate', onPopState);
  document.addEventListener('click', onClick);
  return () => {
    window.removeEventListener('popstate', onPopState);
    document.removeEventListener('click', onClick);
  };
}
