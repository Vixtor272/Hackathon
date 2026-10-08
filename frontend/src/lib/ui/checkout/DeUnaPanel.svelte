<script lang="ts">
  import type { DeUnaQr } from '../../application/usecases';
  import type { PaymentOutcome } from '../../domain';
  import type { DeviceKind } from '../../device';
  import { formatMoney } from '../../format';
  import QrCode from '../shared/QrCode.svelte';

  interface Props {
    amount: number;
    busy: boolean;
    device: DeviceKind;
    qr: DeUnaQr | null;
    onOpenApp: () => void;
    onGenerateQr: () => void;
    onQrOutcome: (outcome: PaymentOutcome) => void;
    onSwitchDevice: (device: DeviceKind) => void;
  }

  let { amount, busy, device, qr, onOpenApp, onGenerateQr, onQrOutcome, onSwitchDevice }: Props = $props();
</script>

<div class="deuna">
  <div class="detected small muted">
    {#if device === 'desktop'}
      💻 Estás en un computador: paga escaneando un código QR con tu celular.
      <button type="button" class="link" onclick={() => onSwitchDevice('mobile')}>Estoy en el celular</button>
    {:else}
      📱 Estás en el celular: te llevamos a DeUna para confirmar el pago.
      <button type="button" class="link" onclick={() => onSwitchDevice('desktop')}>Estoy en un computador</button>
    {/if}
  </div>

  {#if device === 'mobile'}
    <p class="small muted">
      Se abrirá la solicitud de DeUna por {formatMoney(amount)} asociada a este pedido. Si cambias el carrito, la solicitud
      anterior se invalida y se genera otra por el nuevo total.
    </p>
    <button type="button" class="btn btn-deuna" disabled={busy} onclick={onOpenApp}>Pagar con DeUna →</button>
  {:else if !qr}
    <p class="small muted">
      Generaremos un código QR único por {formatMoney(amount)}. Si cambias el carrito, el código anterior deja de valer.
    </p>
    <button type="button" class="btn btn-deuna" disabled={busy} onclick={onGenerateQr}>Generar código QR</button>
  {:else}
    <div class="qr-pay">
      <div class="qr-frame">
        <QrCode value={qr.qrPayload} size={216} label={`Código QR de DeUna por ${formatMoney(qr.payment.amount)}`} />
        <span class="ref mono">{qr.reference}</span>
      </div>
      <div class="steps">
        <strong class="amount">{formatMoney(qr.payment.amount)}</strong>
        <ol class="small">
          <li>Abre la app <strong>DeUna</strong> en tu celular.</li>
          <li>Toca <strong>Pagar con QR</strong> y escanea este código.</li>
          <li>Confirma el pago en la app. Esta página se actualiza sola.</li>
        </ol>
        <div class="sim">
          <span class="small muted">Simulador: representa lo que harías en la app.</span>
          <div class="row">
            <button type="button" class="btn btn-success" disabled={busy} onclick={() => onQrOutcome('approved')}>
              Simular escaneo y pago aprobado
            </button>
            <button type="button" class="btn btn-danger" disabled={busy} onclick={() => onQrOutcome('rejected')}>Simular rechazo</button>
          </div>
          <button type="button" class="link small" disabled={busy} onclick={onGenerateQr}>Generar otro código</button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .deuna {
    display: flex;
    flex-direction: column;
    gap: 10px;
    align-items: flex-start;
  }
  .detected {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
    align-items: baseline;
  }
  .link {
    background: none;
    border: 0;
    padding: 0;
    color: var(--primary);
    font: inherit;
    text-decoration: underline;
    cursor: pointer;
  }
  .link:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .btn-deuna {
    background: #5b21b6;
    border-color: #5b21b6;
    color: #fff;
  }
  .btn-deuna:hover:not(:disabled) {
    background: #4c1d95;
  }
  .qr-pay {
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 24px;
    align-items: center;
    width: 100%;
  }
  .qr-frame {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    padding: 14px;
    border: 2px solid #5b21b6;
    border-radius: 16px;
    background: #fff;
  }
  .ref {
    font-size: 0.85rem;
    letter-spacing: 1px;
    color: #5b21b6;
    font-weight: 700;
  }
  .mono {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  }
  .steps {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .amount {
    font-size: 1.8rem;
  }
  ol {
    margin: 0;
    padding-left: 20px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .sim {
    display: flex;
    flex-direction: column;
    gap: 8px;
    align-items: flex-start;
    border-top: 1px dashed var(--border);
    padding-top: 10px;
  }
  @media (max-width: 720px) {
    .qr-pay {
      grid-template-columns: 1fr;
      justify-items: center;
    }
  }
</style>
