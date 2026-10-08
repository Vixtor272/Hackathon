<script lang="ts">
  import { route, ROUTES } from '../../router';
  import { lastOrderId, lastPaymentId } from '../session';

  const checkoutHref = $derived($lastOrderId ? ROUTES.checkout($lastOrderId) : null);
  const deunaHref = $derived($lastPaymentId ? ROUTES.deuna($lastPaymentId) : null);
</script>

<header class="topnav">
  <a class="brand" href={ROUTES.whatsapp}>
    <span class="logo">💊</span>
    <span>Farmi <span class="muted">· Farmaenlace</span></span>
  </a>
  <nav aria-label="Secciones">
    <a href={ROUTES.whatsapp} class:active={$route.name === 'whatsapp'}>WhatsApp</a>
    {#if checkoutHref}
      <a href={checkoutHref} class:active={$route.name === 'checkout'}>Checkout</a>
    {:else}
      <span class="disabled" title="Se habilita cuando Farmi genere un pedido">Checkout</span>
    {/if}
    {#if deunaHref}
      <a href={deunaHref} class:active={$route.name === 'deuna'}>DeUna</a>
    {:else}
      <span class="disabled" title="Se habilita al iniciar un pago con DeUna">DeUna</span>
    {/if}
  </nav>
  <span class="tag">Experiencia del cliente · datos simulados</span>
</header>

<style>
  .topnav {
    display: flex;
    align-items: center;
    gap: 24px;
    padding: 10px 20px;
    background: var(--surface);
    border-bottom: 1px solid var(--border);
    position: sticky;
    top: 0;
    z-index: 10;
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-weight: 700;
    font-size: 1.05rem;
    color: var(--text);
    text-decoration: none;
  }
  .logo {
    display: inline-grid;
    place-items: center;
    width: 32px;
    height: 32px;
    border-radius: 9px;
    background: var(--primary-soft);
  }
  nav {
    display: flex;
    gap: 4px;
    flex: 1;
    flex-wrap: wrap;
  }
  nav a,
  nav .disabled {
    padding: 6px 12px;
    border-radius: 8px;
    text-decoration: none;
    color: var(--muted);
    font-weight: 600;
  }
  nav a:hover {
    background: var(--surface-muted);
    color: var(--text);
  }
  nav a.active {
    background: var(--primary-soft);
    color: var(--primary-strong);
  }
  nav .disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .tag {
    font-size: 0.75rem;
    color: var(--muted);
    border: 1px solid var(--border);
    border-radius: 999px;
    padding: 3px 10px;
  }
  @media (max-width: 720px) {
    .topnav {
      gap: 12px;
      padding: 8px 12px;
    }
    .tag {
      display: none;
    }
  }
</style>
