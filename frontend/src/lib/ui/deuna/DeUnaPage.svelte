<script lang="ts">
  import { app } from '../../container';
  import {
    errorMessage,
    isApiError,
    PAYMENT_STATUS_LABELS,
    type Bank,
    type PaymentConfirmation,
    type PaymentOutcome,
    type PaymentWithOrder,
  } from '../../domain';
  import { formatMoney } from '../../format';
  import { ROUTES } from '../../router';
  import ErrorBanner from '../shared/ErrorBanner.svelte';
  import StatusPill from '../shared/StatusPill.svelte';
  import { paymentTone } from '../shared/tones';
  import { lastOrderId, lastPaymentId } from '../session';

  interface Props {
    paymentId: string;
  }

  let { paymentId }: Props = $props();

  const { paymentFlow } = app;

  let data = $state<PaymentWithOrder | null>(null);
  let banks = $state<Bank[]>([]);
  let selectedBank = $state<string | null>(null);
  let result = $state<PaymentConfirmation | null>(null);
  let error = $state<string | null>(null);
  let invalidated = $state(false);
  let loading = $state(true);
  let busy = $state(false);

  // Present when the request was opened by scanning the checkout's QR.
  const reference = new URLSearchParams(location.search).get('ref');
  const checkoutHref = $derived(data ? ROUTES.checkout(data.order.id) : ROUTES.whatsapp);
  const pending = $derived(data?.payment.status === 'PENDING' && !result);

  $effect(() => {
    lastPaymentId.set(paymentId);
    void load(paymentId);
  });

  async function load(id: string): Promise<void> {
    loading = true;
    error = null;
    try {
      const [payment, bankList] = await Promise.all([paymentFlow.loadPayment(id), paymentFlow.listBanks()]);
      data = payment;
      banks = bankList;
      lastOrderId.set(payment.order.id);
      invalidated = payment.payment.status === 'INVALIDATED';
    } catch (cause) {
      error = errorMessage(cause);
    } finally {
      loading = false;
    }
  }

  async function confirm(outcome: PaymentOutcome): Promise<void> {
    busy = true;
    error = null;
    try {
      result = await paymentFlow.confirmDeUna(paymentId, outcome, selectedBank);
      data = data ? { ...data, payment: result.payment, order: { ...data.order, status: result.order.status } } : data;
    } catch (cause) {
      error = errorMessage(cause);
      if (isApiError(cause) && cause.code === 'PAYMENT_INVALIDATED') invalidated = true;
    } finally {
      busy = false;
    }
  }
</script>

<div class="deuna">
  <section class="card">
    <header class="head">
      <span class="logo">DeUna</span>
      <span class="muted small">Simulador de pago · bancos ficticios</span>
    </header>

    {#if loading}
      <p class="loading">Cargando solicitud de pago…</p>
    {:else if !data}
      <ErrorBanner message={error ?? 'No se encontró la solicitud de pago.'} />
      <a class="btn" href={ROUTES.whatsapp}>Volver al chat</a>
    {:else}
      <div class="amount">
        <span class="muted">Pedido {data.order.code}{#if reference} · Ref. {reference}{/if}</span>
        <strong>{formatMoney(data.payment.amount)}</strong>
        <StatusPill label={PAYMENT_STATUS_LABELS[result?.payment.status ?? data.payment.status]} tone={paymentTone(result?.payment.status ?? data.payment.status)} />
      </div>

      {#if invalidated}
        <div class="banner banner-warning">
          <div>
            <strong>Esta solicitud ya no es válida.</strong> El carrito cambió o la reserva venció, así que el importe
            anterior quedó invalidado. Vuelve a la página de pago para generar un nuevo enlace por el total vigente.
          </div>
        </div>
        <a class="btn btn-primary" href={checkoutHref}>Volver a la página de pago</a>
      {:else if result}
        {#if result.payment.status === 'APPROVED'}
          <div class="banner banner-success">
            <div>
              <strong>Pago aprobado.</strong> El pedido {result.order.code} quedó <strong>PAGADO</strong> y el inventario se descontó una sola vez.
              Farmi avisará por WhatsApp el tiempo estimado de retiro o entrega.
            </div>
          </div>
        {:else}
          <div class="banner banner-danger">
            <div><strong>Pago rechazado.</strong> Puedes reintentar desde la página de pago mientras la reserva siga vigente.</div>
          </div>
        {/if}
        <a class="btn btn-primary" href={checkoutHref}>Ver el pedido →</a>
      {:else if pending}
        <h2>Elige tu banco</h2>
        <div class="banks">
          {#each banks as bank (bank.id)}
            <button type="button" class="bank" class:selected={selectedBank === bank.id} onclick={() => (selectedBank = bank.id)}>
              <span class="icon">🏦</span>
              <span>{bank.name}</span>
            </button>
          {/each}
        </div>
        <p class="small muted">Se mostrará el mismo importe que aparece en la página de compra: {formatMoney(data.payment.amount)}.</p>
        <ErrorBanner message={error} onclose={() => (error = null)} />
        <div class="row">
          <button type="button" class="btn btn-success" disabled={busy || !selectedBank} onclick={() => void confirm('approved')}>Confirmar pago</button>
          <button type="button" class="btn btn-danger" disabled={busy} onclick={() => void confirm('rejected')}>Rechazar pago</button>
          <a class="btn btn-ghost" href={checkoutHref}>Cancelar y volver</a>
        </div>
      {:else}
        <div class="banner banner-info">
          <div>Esta solicitud ya fue procesada ({PAYMENT_STATUS_LABELS[data.payment.status].toLowerCase()}).</div>
        </div>
        <a class="btn btn-primary" href={checkoutHref}>Ver el pedido →</a>
      {/if}
    {/if}
  </section>
</div>

<style>
  .deuna {
    max-width: 640px;
    margin: 0 auto;
  }
  .head {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
  }
  .logo {
    background: #7c3aed;
    color: #fff;
    font-weight: 800;
    padding: 6px 12px;
    border-radius: 8px;
    letter-spacing: 0.5px;
  }
  .amount {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    margin-bottom: 16px;
  }
  .amount strong {
    font-size: 2rem;
  }
  .banks {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 10px;
    margin-bottom: 12px;
  }
  .bank {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px;
    border: 1px solid var(--border);
    border-radius: 12px;
    background: var(--surface);
    font: inherit;
    font-weight: 600;
    color: inherit;
    cursor: pointer;
    text-align: left;
  }
  .bank:hover {
    background: var(--surface-muted);
  }
  .bank.selected {
    border-color: #7c3aed;
    background: #f3e8ff;
  }
  .icon {
    font-size: 1.3rem;
  }
</style>
