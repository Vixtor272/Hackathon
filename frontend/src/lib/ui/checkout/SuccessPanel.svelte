<script lang="ts">
  import { DELIVERY_STATUS_LABELS, FULFILLMENT_STATUS_LABELS, PAYMENT_METHOD_LABELS, type Order } from '../../domain';
  import { formatDateTime, formatMoney, minutesUntil } from '../../format';
  import StatusPill from '../shared/StatusPill.svelte';
  import { deliveryTone, fulfillmentTone } from '../shared/tones';

  interface Props {
    order: Order;
  }

  let { order }: Props = $props();

  function etaLabel(eta: string | null, fallbackMinutes: number): string {
    const minutes = minutesUntil(eta);
    if (minutes === null) return `≈ ${fallbackMinutes} min`;
    return `≈ ${minutes} min (${formatDateTime(eta)})`;
  }
</script>

<section class="success">
  <div class="banner banner-success">
    <div>
      <strong>¡Pago confirmado!</strong> Pedido {order.code} por {formatMoney(order.total)}
      {#if order.payment} con {PAYMENT_METHOD_LABELS[order.payment.method]}{/if}.
      El inventario se descontó una sola vez y ya notificamos a los puntos de preparación.
    </div>
  </div>

  {#if order.mode === 'pickup'}
    <h2>Retiro en farmacia</h2>
    <div class="grid-2">
      {#each order.fulfillments as fulfillment (fulfillment.pharmacyId)}
        <article class="spot">
          <div class="row between">
            <strong>{fulfillment.pharmacyName}</strong>
            <StatusPill label={FULFILLMENT_STATUS_LABELS[fulfillment.status]} tone={fulfillmentTone(fulfillment.status)} />
          </div>
          <p class="muted small">{fulfillment.address}</p>
          <p>
            {#if fulfillment.status === 'READY'}
              ✅ Listo para recoger.
            {:else if fulfillment.status === 'PICKED_UP'}
              📦 Entregado al cliente.
            {:else}
              ⏱️ Estará listo en {etaLabel(fulfillment.eta, fulfillment.etaMinutes)}.
            {/if}
          </p>
          <ul class="small">
            {#each fulfillment.items as item (item.id)}
              <li>{item.quantity} × {item.medicine} — {item.brand}</li>
            {/each}
          </ul>
        </article>
      {/each}
    </div>
  {:else if order.delivery}
    <h2>Entrega a domicilio</h2>
    <article class="spot">
      <div class="row between">
        <strong>{order.deliveryAddress}</strong>
        <StatusPill label={DELIVERY_STATUS_LABELS[order.delivery.status]} tone={deliveryTone(order.delivery.status)} />
      </div>
      <p class="muted small">Repartidor: {order.delivery.courier.name} · {order.delivery.courier.phone}</p>
      <p>
        {#if order.delivery.status === 'DELIVERED'}
          ✅ Pedido entregado.
        {:else}
          ⏱️ Llegada estimada en {etaLabel(order.delivery.eta, order.delivery.etaMinutes)}.
        {/if}
      </p>
    </article>
  {/if}

  <p class="small muted">Farmi te avisará por WhatsApp con cada cambio de estado.</p>
</section>

<style>
  .success {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .spot {
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 12px 14px;
    background: var(--surface-muted);
  }
  .spot p {
    margin: 6px 0;
  }
  .spot ul {
    margin: 4px 0 0;
    padding-left: 18px;
  }
</style>
