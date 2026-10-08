<script lang="ts">
  import type { Order } from '../../domain';
  import Countdown from '../shared/Countdown.svelte';

  interface Props {
    order: Order;
    busy: boolean;
    onRenew: () => void;
  }

  let { order, busy, onRenew }: Props = $props();
</script>

{#if order.status === 'EXPIRED' || (order.status === 'PENDING' && !order.reservation.active)}
  <div class="banner banner-warning">
    <div class="grow">
      <strong>La reserva venció.</strong> Las unidades se liberaron y el pago pendiente quedó invalidado.
      Renueva la reserva para comprobar disponibilidad y volver a reservar por 10 minutos.
    </div>
    <button type="button" class="btn btn-primary" disabled={busy} onclick={onRenew}>Renovar reserva</button>
  </div>
{:else if order.status === 'PENDING'}
  <div class="banner banner-info">
    <div class="grow">
      <strong>Unidades reservadas.</strong> Tienes este tiempo para confirmar y pagar antes de que se liberen.
    </div>
    <Countdown secondsLeft={order.reservation.secondsLeft} active={order.reservation.active} />
  </div>
{/if}

<style>
  .grow {
    flex: 1;
  }
</style>
