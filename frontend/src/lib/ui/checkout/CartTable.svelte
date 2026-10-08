<script lang="ts">
  import { isEmptyCart, isPaid, type Order, type OrderItem } from '../../domain';
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
  const pending = $derived(!isPaid(order));

  function increaseHint(item: OrderItem): string {
    if (item.requiresPrescription && item.quantity >= item.prescribedQuantity) {
      return `Bajo receta: máximo ${item.prescribedQuantity} ${item.unitLabel} (lo prescrito)`;
    }
    if (!item.canIncrease) return 'Sin stock adicional';
    return item.requiresPrescription ? `Aumentar (hasta ${item.prescribedQuantity} ${item.unitLabel})` : 'Aumentar';
  }
</script>

{#if order.items.length === 0}
  <p class="empty">El carrito está vacío. Agrega productos desde el chat o cancela el pedido.</p>
{:else}
  {#if pending && isEmptyCart(order)}
    <p class="empty">No hay productos por comprar. Usa + para volver a agregar alguno o cancela el pedido.</p>
  {/if}
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
          <tr class:removed={item.quantity === 0}>
            <td>
              <strong>{item.medicine}</strong>
              <div class="muted small">{item.brand} · {item.presentation}</div>
              {#if item.quantity === 0}
                <div class="small">
                  No incluido en el pedido{#if item.canIncrease} · usa + para volver a agregarlo{#if item.requiresPrescription}
                      {' '}(hasta {item.prescribedQuantity} {item.unitLabel}){/if}{/if}
                </div>
              {:else if item.quantity < item.prescribedQuantity}
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
            {#if multiPharmacy}<td class="small" data-label="Local">{item.pharmacyName}</td>{/if}
            <td class="num qty-cell" data-label="Cantidad">
              <div class="qty">
                <button
                  type="button"
                  class="btn btn-icon"
                  aria-label={`Reducir ${item.medicine}`}
                  disabled={!editable || busy || !item.canDecrease}
                  onclick={() => onDecrease(item)}
                >−</button>
                <span class="value">
                  {item.quantity}
                  <span class="muted small">{item.requiresPrescription ? `/ ${item.prescribedQuantity} ` : ''}{item.unitLabel}</span>
                </span>
                <button
                  type="button"
                  class="btn btn-icon"
                  aria-label={`Aumentar ${item.medicine}`}
                  title={increaseHint(item)}
                  disabled={!editable || busy || !item.canIncrease}
                  onclick={() => onIncrease(item)}
                >+</button>
              </div>
            </td>
            <td class="num" data-label="Precio unitario">{formatMoney(item.unitPrice)}</td>
            <td class="num" data-label="Subtotal"><strong>{formatMoney(item.subtotal)}</strong></td>
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
  /* Set aside at zero: greyed out, but + still brings it back. */
  .removed {
    background: var(--surface-muted);
    color: var(--muted);
  }
  .removed td > :global(*):not(.qty) {
    opacity: 0.55;
  }
  .removed strong {
    text-decoration: line-through;
  }
  /* Phones: each product becomes a card so the − / + buttons stay on screen. */
  @media (max-width: 720px) {
    .table thead {
      display: none;
    }
    .table,
    .table tbody,
    .table tr,
    .table td {
      display: block;
      width: 100%;
    }
    .table tr {
      border: 1px solid var(--border);
      border-radius: 12px;
      padding: 10px 12px;
      margin-bottom: 10px;
    }
    .table td {
      border: 0;
      padding: 4px 0;
      text-align: left;
    }
    .table td[data-label] {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 12px;
    }
    .table td[data-label]::before {
      content: attr(data-label);
      color: var(--muted);
      font-size: 0.85rem;
      font-weight: 400;
    }
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
