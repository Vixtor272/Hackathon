<script lang="ts">
  import type { Order, OrderItem } from '../../domain';
  import { formatMoney } from '../../format';
  import StatusPill from '../shared/StatusPill.svelte';

  interface Props {
    order: Order;
    editable: boolean;
    busy: boolean;
    onIncrease: (item: OrderItem) => void;
    onDecrease: (item: OrderItem) => void;
  }

  let { order, editable, busy, onIncrease, onDecrease }: Props = $props();

  const multiPharmacy = $derived(new Set(order.items.map((item) => item.pharmacyId)).size > 1);
</script>

{#if order.items.length === 0}
  <p class="empty">El carrito está vacío. Agrega productos desde el chat o cancela el pedido.</p>
{:else}
  <div class="scroll">
    <table class="table">
      <thead>
        <tr>
          <th>Producto</th>
          <th>Condición</th>
          {#if multiPharmacy}<th>Local</th>{/if}
          <th class="num">Cantidad</th>
          <th class="num">Precio unitario</th>
          <th class="num">Subtotal</th>
        </tr>
      </thead>
      <tbody>
        {#each order.items as item (item.id)}
          <tr>
            <td>
              <strong>{item.medicine}</strong>
              <div class="muted small">{item.brand} · {item.presentation}</div>
              {#if item.quantity < item.prescribedQuantity}
                <div class="small diff">
                  Prescritas: {item.prescribedQuantity} {item.unitLabel} · faltan {item.prescribedQuantity - item.quantity}
                </div>
              {/if}
            </td>
            <td>
              {#if item.requiresPrescription}
                <StatusPill label="Bajo receta" tone="warning" />
              {:else}
                <StatusPill label="Venta libre" tone="success" />
              {/if}
            </td>
            {#if multiPharmacy}<td class="small">{item.pharmacyName}</td>{/if}
            <td class="num">
              <div class="qty">
                <button
                  type="button"
                  class="btn btn-icon"
                  aria-label={`Reducir ${item.medicine}`}
                  disabled={!editable || busy || !item.canDecrease}
                  onclick={() => onDecrease(item)}
                >−</button>
                <span class="value">{item.quantity} <span class="muted small">{item.unitLabel}</span></span>
                <button
                  type="button"
                  class="btn btn-icon"
                  aria-label={`Aumentar ${item.medicine}`}
                  title={item.requiresPrescription ? 'Bajo receta: no se puede aumentar desde esta página' : !item.canIncrease ? 'Sin stock adicional' : 'Aumentar'}
                  disabled={!editable || busy || !item.canIncrease}
                  onclick={() => onIncrease(item)}
                >+</button>
              </div>
            </td>
            <td class="num">{formatMoney(item.unitPrice)}</td>
            <td class="num"><strong>{formatMoney(item.subtotal)}</strong></td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<div class="totals">
  <div class="row between"><span>Productos</span><span>{formatMoney(order.subtotal)}</span></div>
  {#if order.mode === 'delivery'}
    <div class="row between"><span>Envío</span><span>{formatMoney(order.deliveryFee)}</span></div>
  {/if}
  <div class="row between total"><span>Total</span><span>{formatMoney(order.total)}</span></div>
</div>

<style>
  .scroll {
    overflow-x: auto;
  }
  .qty {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }
  .value {
    min-width: 72px;
    text-align: center;
    font-weight: 600;
  }
  .diff {
    color: var(--warning);
  }
  .totals {
    margin-top: 14px;
    margin-left: auto;
    max-width: 320px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .total {
    border-top: 1px solid var(--border);
    padding-top: 8px;
    margin-top: 4px;
    font-size: 1.15rem;
    font-weight: 700;
  }
</style>
