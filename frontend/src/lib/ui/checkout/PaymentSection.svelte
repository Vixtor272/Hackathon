<script lang="ts">
  import type { DeUnaQr } from '../../application/usecases';
  import type { CardForm as CardDetails, PaymentMethod, PaymentOption, PaymentOutcome } from '../../domain';
  import type { DeviceKind } from '../../device';
  import CardForm from './CardForm.svelte';
  import DeUnaPanel from './DeUnaPanel.svelte';

  interface Props {
    options: PaymentOption[];
    amount: number;
    busy: boolean;
    device: DeviceKind;
    deunaQr: DeUnaQr | null;
    onCard: (card: CardDetails) => void;
    onDeUna: () => void;
    onDeUnaQr: () => void;
    onDeUnaQrOutcome: (outcome: PaymentOutcome) => void;
    onSwitchDevice: (device: DeviceKind) => void;
  }

  let { options, amount, busy, device, deunaQr, onCard, onDeUna, onDeUnaQr, onDeUnaQrOutcome, onSwitchDevice }: Props = $props();
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
    <CardForm {amount} {busy} onSubmit={onCard} />
  {:else if selected === 'deuna'}
    <DeUnaPanel
      {amount}
      {busy}
      {device}
      qr={deunaQr}
      onOpenApp={onDeUna}
      onGenerateQr={onDeUnaQr}
      onQrOutcome={onDeUnaQrOutcome}
      {onSwitchDevice}
    />
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
</style>
