<script lang="ts">
  import { ORDER_STATUS_LABELS, type Order } from '../../domain';
  import { formatMoney } from '../../format';
  import StatusPill from '../shared/StatusPill.svelte';
  import { orderTone } from '../shared/tones';

  interface Props {
    order: Order;
  }

  let { order }: Props = $props();

  const pharmacies = $derived(
    order.fulfillments.length > 0
      ? order.fulfillments
      : [...new Map(order.items.filter((item) => item.quantity > 0).map((item) => [item.pharmacyId, { pharmacyId: item.pharmacyId, pharmacyName: item.pharmacyName, address: '' }])).values()],
  );
</script>

<header class="summary">
  <div class="row between">
    <div>
      <h1>Pedido {order.code}</h1>
      <p class="muted">{order.clientName} · {order.phone} · {order.zone.label}</p>
    </div>
    <StatusPill label={ORDER_STATUS_LABELS[order.status]} tone={orderTone(order.status)} />
  </div>

  {#if order.mode === 'pickup'}
    <div class="mode">
      <strong>🏪 Retiro en farmacia</strong>
      <ul>
        {#each pharmacies as pharmacy (pharmacy.pharmacyId)}
          <li>{pharmacy.pharmacyName}{#if pharmacy.address} — <span class="muted">{pharmacy.address}</span>{/if}</li>
        {/each}
      </ul>
      {#if pharmacies.length > 1}
        <p class="small muted">La receta se reparte entre {pharmacies.length} locales; cada producto indica dónde se retira.</p>
      {/if}
    </div>
  {:else}
    <div class="mode">
      <strong>🛵 Entrega a domicilio</strong>
      <p>{order.deliveryAddress ?? 'Dirección pendiente'}</p>
      <p class="small muted">
        Envío {formatMoney(order.deliveryFee)}
        {#if order.delivery?.courier} · Repartidor: {order.delivery.courier.name}{/if}
      </p>
    </div>
  {/if}
</header>

<style>
  .summary {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .mode {
    background: var(--surface-muted);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 10px 14px;
  }
  .mode ul {
    margin: 6px 0 0;
    padding-left: 18px;
  }
  .mode p {
    margin: 4px 0 0;
  }
</style>
