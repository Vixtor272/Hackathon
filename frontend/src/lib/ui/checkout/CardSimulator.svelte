<script lang="ts">
  import type { PaymentOutcome } from '../../domain';
  import { formatMoney } from '../../format';

  interface Props {
    amount: number;
    busy: boolean;
    onOutcome: (outcome: PaymentOutcome) => void;
  }

  let { amount, busy, onOutcome }: Props = $props();
</script>

<div class="simulator">
  <div class="fake-card" aria-hidden="true">
    <span class="chip"></span>
    <span class="number">•••• •••• •••• 4242</span>
    <span class="meta"><span>TITULAR DEMO</span><span>12/30</span></span>
  </div>
  <div class="actions">
    <p class="small muted">Simulador sin datos bancarios reales. El resultado se registra como un pago de {formatMoney(amount)}.</p>
    <div class="row">
      <button type="button" class="btn btn-success" disabled={busy} onclick={() => onOutcome('approved')}>Simular pago aprobado</button>
      <button type="button" class="btn btn-danger" disabled={busy} onclick={() => onOutcome('rejected')}>Simular pago rechazado</button>
    </div>
  </div>
</div>

<style>
  .simulator {
    display: grid;
    grid-template-columns: 260px 1fr;
    gap: 18px;
    align-items: center;
  }
  .fake-card {
    aspect-ratio: 1.6;
    border-radius: 14px;
    background: linear-gradient(135deg, #0f766e, #134e4a 60%, #1e293b);
    color: #fff;
    padding: 16px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    box-shadow: var(--shadow);
  }
  .chip {
    width: 36px;
    height: 26px;
    border-radius: 6px;
    background: linear-gradient(135deg, #fde68a, #d97706);
  }
  .number {
    font-size: 1.1rem;
    letter-spacing: 2px;
    font-variant-numeric: tabular-nums;
  }
  .meta {
    display: flex;
    justify-content: space-between;
    font-size: 0.75rem;
    opacity: 0.85;
  }
  @media (max-width: 720px) {
    .simulator {
      grid-template-columns: 1fr;
    }
  }
</style>
