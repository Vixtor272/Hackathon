<script lang="ts">
  import type { DeUnaQr } from '../../application/usecases';
  import { app } from '../../container';
  import { detectDevice, type DeviceKind } from '../../device';
  import {
    errorMessage,
    isApiError,
    isEmptyCart,
    isPaid,
    maskedNumber,
    type CardForm,
    type Order,
    type OrderItem,
    type PaymentOptions,
    type PaymentOutcome,
  } from '../../domain';
  import { openLink, ROUTES } from '../../router';
  import ErrorBanner from '../shared/ErrorBanner.svelte';
  import { startPolling } from '../shared/poll';
  import { lastOrderId, lastPaymentId } from '../session';
  import CartTable from './CartTable.svelte';
  import OrderSummary from './OrderSummary.svelte';
  import PaymentSection from './PaymentSection.svelte';
  import ReservationBanner from './ReservationBanner.svelte';
  import SuccessPanel from './SuccessPanel.svelte';

  interface Props {
    orderId: string;
  }

  let { orderId }: Props = $props();

  const { checkout, paymentFlow } = app;
  const POLL_MS = 3000;

  let order = $state<Order | null>(null);
  let options = $state<PaymentOptions | null>(null);
  let loadError = $state<string | null>(null);
  let actionError = $state<string | null>(null);
  let notice = $state<{ text: string; tone: 'info' | 'warning' } | null>(null);
  let busy = $state(false);
  let device = $state<DeviceKind>(detectDevice());
  let deunaQr = $state<DeUnaQr | null>(null);

  const paid = $derived(order ? isPaid(order) : false);
  const editable = $derived(order ? order.status === 'PENDING' && order.reservation.active : false);

  // A QR is only shown while its payment is the order's pending one: a cart
  // change, a rejection or an expiry voids it.
  const activeQr = $derived(
    deunaQr && order?.payment?.id === deunaQr.payment.id && order.payment.status === 'PENDING' ? deunaQr : null,
  );

  $effect(() => {
    lastOrderId.set(orderId);
    return startPolling(refresh, POLL_MS);
  });

  async function refresh(): Promise<void> {
    try {
      order = await checkout.load(orderId);
      loadError = null;
      if (!options && order.status === 'PENDING') {
        options = await checkout.loadPaymentOptions(orderId);
      }
    } catch (cause) {
      loadError = errorMessage(cause);
    }
  }

  async function run(action: () => Promise<Order | null>): Promise<void> {
    busy = true;
    actionError = null;
    try {
      const result = await action();
      if (result) order = result;
    } catch (cause) {
      actionError = errorMessage(cause);
      if (isApiError(cause) && (cause.code === 'RESERVATION_EXPIRED' || cause.code === 'INVALID_STATE' || cause.code === 'PAYMENT_INVALIDATED')) {
        await refresh();
      }
    } finally {
      busy = false;
    }
  }

  function increase(item: OrderItem): void {
    if (!order) return;
    const current = order;
    void run(() => checkout.increase(current, item));
  }

  function decrease(item: OrderItem): void {
    if (!order) return;
    const current = order;
    void run(() => checkout.decrease(current, item));
  }

  function cancel(): void {
    if (!window.confirm('¿Cancelar el pedido y liberar las unidades reservadas?')) return;
    void run(async () => {
      const result = await checkout.cancel(orderId);
      notice = { text: 'Pedido cancelado. Las unidades reservadas quedaron disponibles de nuevo.', tone: 'info' };
      return result;
    });
  }

  function renew(): void {
    void run(async () => {
      const result = await checkout.renew(orderId);
      notice = { text: 'Reserva renovada por 10 minutos.', tone: 'info' };
      options = null;
      await refresh();
      return result;
    });
  }

  function payWithCard(card: CardForm): void {
    void run(async () => {
      const result = await paymentFlow.payWithCardDetails(orderId, card);
      lastPaymentId.set(result.payment.id);
      notice =
        result.payment.status === 'APPROVED'
          ? null
          : {
              text: `La tarjeta ${maskedNumber(card.number)} fue rechazada (fondos insuficientes). Prueba con otra tarjeta mientras la reserva siga vigente.`,
              tone: 'warning',
            };
      return result.order;
    });
  }

  function payWithDeUnaQr(): void {
    void run(async () => {
      const qr = await paymentFlow.startDeUnaQr(orderId);
      lastPaymentId.set(qr.payment.id);
      deunaQr = qr;
      return checkout.load(orderId); // the order now points at the new pending payment
    });
  }

  function confirmDeUnaQr(outcome: PaymentOutcome): void {
    const qr = activeQr;
    if (!qr) return;
    void run(async () => {
      const result = await paymentFlow.confirmDeUnaQr(qr.payment.id, outcome);
      notice =
        outcome === 'approved' ? null : { text: 'DeUna rechazó el pago. Genera un nuevo código QR o elige otro método.', tone: 'warning' };
      return result.order;
    });
  }

  function payWithDeUna(): void {
    void run(async () => {
      const { payment, link } = await paymentFlow.startDeUna(orderId);
      lastPaymentId.set(payment.id);
      openLink(link);
      return null;
    });
  }
</script>

<div class="checkout">
  {#if !order && loadError}
    <ErrorBanner message={loadError} />
    <a class="btn" href={ROUTES.whatsapp}>Volver al chat</a>
  {:else if !order}
    <p class="loading">Cargando pedido…</p>
  {:else}
    <section class="card">
      <OrderSummary {order} />
    </section>

    <ErrorBanner message={loadError} tone="warning" />

    {#if !paid && order.status !== 'CANCELLED'}
      <ReservationBanner {order} {busy} onRenew={renew} />
    {/if}

    {#if order.status === 'CANCELLED'}
      <div class="banner banner-danger"><div>Este pedido fue cancelado. Vuelve al chat para iniciar una nueva compra.</div></div>
    {/if}

    <ErrorBanner message={notice?.text ?? null} tone={notice?.tone ?? 'info'} onclose={() => (notice = null)} />
    <ErrorBanner message={actionError} onclose={() => (actionError = null)} />

    <section class="card">
      <h2>Carrito</h2>
      {#if !paid && order.status !== 'CANCELLED'}
        <p class="small muted">
          Los productos de venta libre se pueden aumentar o reducir según stock. Los medicamentos bajo receta se pueden reducir y
          volver a subir hasta la cantidad prescrita, nunca más. Al llegar a cero, el producto queda en gris y no se cobra; con + se vuelve a agregar.
        </p>
      {/if}
      <CartTable {order} {editable} {busy} onIncrease={increase} onDecrease={decrease} />
    </section>

    {#if paid}
      <section class="card">
        <SuccessPanel {order} />
      </section>
    {:else if checkout.canPay(order) && options}
      <section class="card">
        <PaymentSection
          options={options.options}
          amount={order.total}
          {busy}
          {device}
          deunaQr={activeQr}
          onCard={payWithCard}
          onDeUna={payWithDeUna}
          onDeUnaQr={payWithDeUnaQr}
          onDeUnaQrOutcome={confirmDeUnaQr}
          onSwitchDevice={(next) => (device = next)}
        />
      </section>
    {:else if order.status === 'PENDING' && isEmptyCart(order)}
      <div class="banner banner-warning"><div>El carrito está vacío: no se puede pagar.</div></div>
    {/if}

    <footer class="row between">
      <a class="btn" href={ROUTES.whatsapp}>← Volver al chat</a>
      {#if !paid && order.status !== 'CANCELLED'}
        <button type="button" class="btn btn-danger" disabled={busy} onclick={cancel}>Cancelar pedido</button>
      {/if}
    </footer>
  {/if}
</div>

<style>
  .checkout {
    display: flex;
    flex-direction: column;
    gap: 16px;
    max-width: 960px;
    margin: 0 auto;
  }
  .checkout .card + .card {
    margin-top: 0;
  }
</style>
