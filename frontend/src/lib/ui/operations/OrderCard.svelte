<script lang="ts">
  import {
    DELIVERY_STATUS_LABELS,
    FULFILLMENT_STATUS_LABELS,
    ORDER_STATUS_LABELS,
    type Order,
  } from '../../domain';
  import { app } from '../../container';
  import { formatDateTime, formatMoney } from '../../format';
  import StatusPill from '../shared/StatusPill.svelte';
  import { deliveryTone, fulfillmentTone, orderTone } from '../shared/tones';

  export type BoardContext = { kind: 'pharmacy'; pharmacyId: string } | { kind: 'courier' };

  interface Props {
    order: Order;
    context: BoardContext;
    busy: boolean;
    onAdvance: (order: Order) => void;
  }

  let { order, context, busy, onAdvance }: Props = $props();

  const fulfillment = $derived(
    context.kind === 'pharmacy' ? order.fulfillments.find((f) => f.pharmacyId === context.pharmacyId) ?? null : null,
  );
  const items = $derived(fulfillment ? fulfillment.items : order.items.map((item) => ({ id: item.id, medicine: item.medicine, brand: item.brand, quantity: item.quantity })));
  const step = $derived(
    context.kind === 'pharmacy'
      ? fulfillment
        ? app.operations.nextPharmacyStep(fulfillment.status)
        : null
      : order.delivery
        ? app.operations.nextDeliveryStep(order.delivery.status)
        : null,
  );
</script>

<article class="order">
  <header class="row between">
    <div>
      <strong>{order.code}</strong>
      <span class="muted small"> · {order.clientName} · {order.phone}</span>
    </div>
    <StatusPill label={ORDER_STATUS_LABELS[order.status]} tone={orderTone(order.status)} />
  </header>

  <ul class="items small">
    {#each items as item (item.id)}
      <li>{item.quantity} × {item.medicine} — {item.brand}</li>
    {/each}
  </ul>

  <div class="meta small muted">
    {#if order.mode === 'pickup'}
      Retiro en farmacia{#if order.fulfillments.length > 1} · pedido repartido en {order.fulfillments.length} locales{/if}
    {:else}
      Domicilio: {order.deliveryAddress}
    {/if}
    · Total {formatMoney(order.total)} · Pagado {formatDateTime(order.paidAt)}
  </div>

  <footer class="row between">
    {#if context.kind === 'pharmacy' && fulfillment}
      <StatusPill label={FULFILLMENT_STATUS_LABELS[fulfillment.status]} tone={fulfillmentTone(fulfillment.status)} />
    {:else if order.delivery}
      <StatusPill label={DELIVERY_STATUS_LABELS[order.delivery.status]} tone={deliveryTone(order.delivery.status)} />
      <span class="small muted">{order.delivery.courier.name}</span>
    {/if}
    {#if step}
      <button type="button" class="btn btn-primary btn-sm" disabled={busy} onclick={() => onAdvance(order)}>Marcar «{step.label}»</button>
    {:else}
      <span class="small muted">Sin acciones pendientes</span>
    {/if}
  </footer>
</article>

<style>
  .order {
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 12px 14px;
    background: var(--surface);
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .items {
    margin: 0;
    padding-left: 18px;
  }
</style>
