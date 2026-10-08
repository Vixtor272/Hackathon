<script lang="ts">
  import type { PaymentMethod, PaymentOption, PaymentOutcome } from '../../domain';
  import { formatMoney } from '../../format';
  import CardSimulator from './CardSimulator.svelte';

  interface Props {
    options: PaymentOption[];
    amount: number;
    busy: boolean;
    onCard: (outcome: PaymentOutcome) => void;
    onDeUna: () => void;
  }

  let { options, amount, busy, onCard, onDeUna }: Props = $props();
  let selected = $state<PaymentMethod | null>(null);
</script>

<section class="payment">
  <h2>Elige cómo pagar</h2>
  <div class="methods">
    {#each options as option (option.method)}
      <button
        type="button"
        class="method"
        class:selected={selected === option.method}
        onclick={() => (selected = option.method)}
      >
        <span class="icon">{option.method === 'card' ? '💳' : '🏦'}</span>
        <span class="body">
          <strong>{option.label}</strong>
          <span class="muted small">{option.description}</span>
        </span>
      </button>
    {/each}
  </div>

  {#if selected === 'card'}
    <CardSimulator {amount} {busy} onOutcome={onCard} />
  {:else if selected === 'deuna'}
    <div class="deuna">
      <p class="small muted">
        Se generará un enlace de prueba por {formatMoney(amount)} asociado a este pedido. Si cambias el carrito, el
        enlace anterior se invalida y se genera otro por el nuevo total.
      </p>
      <button type="button" class="btn btn-primary" disabled={busy} onclick={onDeUna}>Pagar con DeUna →</button>
    </div>
  {:else}
    <p class="small muted">Selecciona un método para continuar.</p>
  {/if}
</section>

<style>
  .payment {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .methods {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
    gap: 10px;
  }
  .method {
    display: flex;
    gap: 12px;
    align-items: center;
    padding: 12px 14px;
    border: 1px solid var(--border);
    border-radius: 12px;
    background: var(--surface);
    cursor: pointer;
    text-align: left;
    font: inherit;
    color: inherit;
  }
  .method:hover {
    background: var(--surface-muted);
  }
  .method.selected {
    border-color: var(--primary);
    background: var(--primary-soft);
  }
  .icon {
    font-size: 1.5rem;
  }
  .body {
    display: flex;
    flex-direction: column;
  }
  .deuna {
    display: flex;
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
  }
</style>
