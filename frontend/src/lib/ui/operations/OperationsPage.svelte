<script lang="ts">
  import { app } from '../../container';
  import { errorMessage, type Notification, type Order, type Pharmacy } from '../../domain';
  import ErrorBanner from '../shared/ErrorBanner.svelte';
  import { startPolling } from '../shared/poll';
  import NotificationsTable from './NotificationsTable.svelte';
  import OrderCard, { type BoardContext } from './OrderCard.svelte';

  const { operations } = app;
  const POLL_MS = 3000;
  const COURIER = 'courier';
  const NOTIFICATIONS = 'notifications';

  let pharmacies = $state<Pharmacy[]>([]);
  let activeKey = $state<string>('');
  let orders = $state<Order[]>([]);
  let notifications = $state<Notification[]>([]);
  let error = $state<string | null>(null);
  let busy = $state(false);

  const activePharmacy = $derived(pharmacies.find((pharmacy) => `pharmacy:${pharmacy.id}` === activeKey) ?? null);
  const context = $derived<BoardContext | null>(
    activeKey === COURIER ? { kind: 'courier' } : activePharmacy ? { kind: 'pharmacy', pharmacyId: activePharmacy.id } : null,
  );

  $effect(() => {
    void operations
      .listAllPharmacies()
      .then((list) => {
        pharmacies = list;
        const first = list[0];
        if (activeKey === '' && first) activeKey = `pharmacy:${first.id}`;
      })
      .catch((cause: unknown) => {
        error = errorMessage(cause);
      });
  });

  $effect(() => {
    const key = activeKey;
    const ctx = context;
    if (key === '') return;
    return startPolling(() => load(key, ctx), POLL_MS);
  });

  async function load(key: string, ctx: BoardContext | null): Promise<void> {
    try {
      if (key === NOTIFICATIONS) {
        notifications = await operations.loadNotifications();
      } else if (ctx?.kind === 'courier') {
        orders = await operations.ordersForCourier();
      } else if (ctx?.kind === 'pharmacy') {
        orders = await operations.ordersForPharmacy(ctx.pharmacyId);
      }
      error = null;
    } catch (cause) {
      error = errorMessage(cause);
    }
  }

  async function advance(order: Order): Promise<void> {
    if (!context) return;
    busy = true;
    error = null;
    try {
      let updated: Order;
      if (context.kind === 'pharmacy') {
        const fulfillment = order.fulfillments.find((f) => f.pharmacyId === (context.kind === 'pharmacy' ? context.pharmacyId : ''));
        if (!fulfillment) return;
        updated = await operations.advancePharmacy(order.id, fulfillment.pharmacyId, fulfillment.status);
      } else {
        if (!order.delivery) return;
        updated = await operations.advanceDelivery(order.id, order.delivery.status);
      }
      orders = orders.map((candidate) => (candidate.id === updated.id ? updated : candidate));
    } catch (cause) {
      error = errorMessage(cause);
    } finally {
      busy = false;
    }
  }

  function select(key: string): void {
    if (key !== activeKey) {
      orders = [];
      activeKey = key;
    }
  }
</script>

<div class="operations">
  <header>
    <h1>Panel de operaciones</h1>
    <p class="muted">Cajeros de farmacia, repartidor y bandeja de notificaciones simuladas. Se actualiza cada 3 segundos.</p>
  </header>

  <nav class="tabs" aria-label="Puntos de operación">
    {#each pharmacies as pharmacy (pharmacy.id)}
      <button type="button" class="tab" class:active={activeKey === `pharmacy:${pharmacy.id}`} onclick={() => select(`pharmacy:${pharmacy.id}`)}>
        🏪 {pharmacy.name}
      </button>
    {/each}
    <button type="button" class="tab" class:active={activeKey === COURIER} onclick={() => select(COURIER)}>🛵 Reparto</button>
    <button type="button" class="tab" class:active={activeKey === NOTIFICATIONS} onclick={() => select(NOTIFICATIONS)}>🔔 Notificaciones</button>
  </nav>

  <ErrorBanner message={error} onclose={() => (error = null)} />

  <section class="card">
    {#if activeKey === NOTIFICATIONS}
      <h2>Notificaciones enviadas</h2>
      <NotificationsTable {notifications} />
    {:else if context}
      <div class="row between">
        <h2>
          {#if context.kind === 'courier'}
            Pedidos a domicilio
          {:else if activePharmacy}
            {activePharmacy.name} <span class="muted small">· {activePharmacy.chain} · {activePharmacy.address}</span>
          {/if}
        </h2>
        <span class="muted small">{orders.length} pedido{orders.length === 1 ? '' : 's'}</span>
      </div>
      {#if orders.length === 0}
        <p class="empty">No hay pedidos pagados para este punto todavía.</p>
      {:else}
        <div class="list">
          {#each orders as order (order.id)}
            <OrderCard {order} {context} {busy} onAdvance={(target) => void advance(target)} />
          {/each}
        </div>
      {/if}
    {:else}
      <p class="loading">Cargando puntos de operación…</p>
    {/if}
  </section>
</div>

<style>
  .operations {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .tab {
    padding: 8px 14px;
    border-radius: 999px;
    border: 1px solid var(--border);
    background: var(--surface);
    font: inherit;
    font-weight: 600;
    color: var(--muted);
    cursor: pointer;
  }
  .tab:hover {
    background: var(--surface-muted);
  }
  .tab.active {
    border-color: var(--primary);
    background: var(--primary-soft);
    color: var(--primary-strong);
  }
  .list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 12px;
  }
</style>
