<script lang="ts">
  import {
    CARD_BRAND_LABELS,
    cvvLength,
    detectBrand,
    digitsOf,
    EMPTY_CARD,
    formatCardNumber,
    formatExpiry,
    hasErrors,
    TEST_CARDS,
    validateCard,
    type CardField,
    type CardForm,
  } from '../../domain';
  import { formatMoney } from '../../format';

  interface Props {
    amount: number;
    busy: boolean;
    onSubmit: (card: CardForm) => void;
  }

  let { amount, busy, onSubmit }: Props = $props();

  let card = $state<CardForm>({ ...EMPTY_CARD });
  let touched = $state<Partial<Record<CardField, boolean>>>({});
  let submitted = $state(false);
  let flipped = $state(false);

  const brand = $derived(detectBrand(card.number));
  const errors = $derived(validateCard(card));
  const shown = (field: CardField): string | undefined => (submitted || touched[field] ? errors[field] : undefined);

  const previewNumber = $derived.by(() => {
    const typed = formatCardNumber(card.number);
    const template = brand === 'amex' ? '•••• •••••• •••••' : '•••• •••• •••• ••••';
    return typed.length >= template.length ? typed : typed + template.slice(typed.length);
  });

  function onNumber(event: Event): void {
    card.number = formatCardNumber((event.currentTarget as HTMLInputElement).value);
  }

  function onExpiry(event: Event): void {
    const input = event.currentTarget as HTMLInputElement;
    // Let backspace remove the slash instead of re-adding it.
    const raw = input.value.endsWith('/') && input.value.length < card.expiry.length ? input.value.slice(0, -1) : input.value;
    card.expiry = formatExpiry(raw);
  }

  function onCvv(event: Event): void {
    card.cvv = digitsOf((event.currentTarget as HTMLInputElement).value).slice(0, cvvLength(brand));
  }

  function useTestCard(number: string): void {
    const cvv = detectBrand(number) === 'amex' ? '1234' : '123';
    card = { number, holder: card.holder || 'María Pérez', expiry: card.expiry || '12/30', cvv };
    touched = {};
    submitted = false;
  }

  function submit(event: SubmitEvent): void {
    event.preventDefault();
    submitted = true;
    if (hasErrors(errors)) return;
    onSubmit({ ...card });
  }
</script>

<div class="card-pay">
  <div class="preview" class:flipped aria-hidden="true">
    <div class={`face front brand-${brand}`}>
      <div class="top">
        <span class="chip"></span>
        <span class="brand">{CARD_BRAND_LABELS[brand]}</span>
      </div>
      <span class="number">{previewNumber}</span>
      <div class="meta">
        <span>
          <small>Titular</small>
          {card.holder.trim().toUpperCase() || 'NOMBRE APELLIDO'}
        </span>
        <span>
          <small>Vence</small>
          {card.expiry || 'MM/AA'}
        </span>
      </div>
    </div>
    <div class={`face back brand-${brand}`}>
      <span class="stripe"></span>
      <span class="cvv-box">{card.cvv || '•'.repeat(cvvLength(brand))}</span>
    </div>
  </div>

  <form class="form" novalidate onsubmit={submit}>
    <div class="banner banner-info small">
      <div>Simulador: no ingreses datos reales. Los datos de la tarjeta no salen de esta página.</div>
    </div>

    <label class="field">
      <span>Número de tarjeta</span>
      <input
        class="input mono"
        class:invalid={shown('number')}
        inputmode="numeric"
        autocomplete="off"
        placeholder="1234 5678 9012 3456"
        value={card.number}
        oninput={onNumber}
        onblur={() => (touched.number = true)}
        aria-invalid={Boolean(shown('number'))}
      />
      {#if shown('number')}<small class="error">{shown('number')}</small>{/if}
    </label>

    <label class="field">
      <span>Nombre del titular</span>
      <input
        class="input"
        class:invalid={shown('holder')}
        autocomplete="off"
        placeholder="Como aparece en la tarjeta"
        bind:value={card.holder}
        onblur={() => (touched.holder = true)}
        aria-invalid={Boolean(shown('holder'))}
      />
      {#if shown('holder')}<small class="error">{shown('holder')}</small>{/if}
    </label>

    <div class="pair">
      <label class="field">
        <span>Vencimiento</span>
        <input
          class="input mono"
          class:invalid={shown('expiry')}
          inputmode="numeric"
          autocomplete="off"
          placeholder="MM/AA"
          value={card.expiry}
          oninput={onExpiry}
          onblur={() => (touched.expiry = true)}
          aria-invalid={Boolean(shown('expiry'))}
        />
        {#if shown('expiry')}<small class="error">{shown('expiry')}</small>{/if}
      </label>
      <label class="field">
        <span>CVV</span>
        <input
          class="input mono"
          class:invalid={shown('cvv')}
          inputmode="numeric"
          autocomplete="off"
          type="password"
          placeholder={'•'.repeat(cvvLength(brand))}
          value={card.cvv}
          oninput={onCvv}
          onfocus={() => (flipped = true)}
          onblur={() => {
            flipped = false;
            touched.cvv = true;
          }}
          aria-invalid={Boolean(shown('cvv'))}
        />
        {#if shown('cvv')}<small class="error">{shown('cvv')}</small>{/if}
      </label>
    </div>

    <button type="submit" class="btn btn-primary pay" disabled={busy}>
      {busy ? 'Procesando…' : `Pagar ${formatMoney(amount)}`}
    </button>

    <details class="tests">
      <summary class="small">Tarjetas de prueba</summary>
      <ul>
        {#each TEST_CARDS as test (test.number)}
          <li>
            <button type="button" class="btn btn-sm" onclick={() => useTestCard(test.number)}>Usar</button>
            <span class="mono small">{test.number}</span>
            <span class="small" class:approved={test.outcome === 'approved'} class:rejected={test.outcome === 'rejected'}>{test.label}</span>
          </li>
        {/each}
      </ul>
      <p class="small muted">Cualquier otro número válido se aprueba. Vencimiento futuro y CVV de 3 dígitos (4 en American Express).</p>
    </details>
  </form>
</div>

<style>
  .card-pay {
    display: grid;
    grid-template-columns: 300px 1fr;
    gap: 24px;
    align-items: start;
  }
  .preview {
    position: relative;
    aspect-ratio: 1.586;
    perspective: 900px;
  }
  .face {
    position: absolute;
    inset: 0;
    border-radius: 14px;
    color: #fff;
    padding: 18px;
    box-shadow: var(--shadow);
    backface-visibility: hidden;
    transition: transform 0.5s ease;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    background: linear-gradient(135deg, #0f766e, #134e4a 60%, #1e293b);
  }
  .back {
    transform: rotateY(180deg);
    padding: 18px 0;
    justify-content: flex-start;
    gap: 18px;
  }
  .flipped .front {
    transform: rotateY(-180deg);
  }
  .flipped .back {
    transform: rotateY(0deg);
  }
  .brand-visa {
    background: linear-gradient(135deg, #1a3fa8, #0b2470 65%, #0a1a4a);
  }
  .brand-mastercard {
    background: linear-gradient(135deg, #2b2b2b, #111 60%, #3a1d00);
  }
  .brand-amex {
    background: linear-gradient(135deg, #2e7dbf, #1f5f96 60%, #154470);
  }
  .brand-diners {
    background: linear-gradient(135deg, #5b6470, #39414b 60%, #22272e);
  }
  .brand-discover {
    background: linear-gradient(135deg, #e86a1c, #b44c0b 60%, #4a2306);
  }
  .top {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .chip {
    width: 40px;
    height: 30px;
    border-radius: 6px;
    background: linear-gradient(135deg, #fde68a, #d97706);
  }
  .brand {
    font-weight: 800;
    letter-spacing: 0.5px;
    font-style: italic;
  }
  .number {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    font-size: 1.15rem;
    letter-spacing: 1.5px;
    white-space: nowrap;
  }
  .meta {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    font-size: 0.8rem;
    text-transform: uppercase;
  }
  .meta span {
    display: flex;
    flex-direction: column;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .meta small {
    font-size: 0.6rem;
    opacity: 0.75;
  }
  .stripe {
    display: block;
    height: 40px;
    background: #111;
  }
  .cvv-box {
    align-self: flex-end;
    margin-right: 18px;
    background: #fff;
    color: #111;
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    padding: 6px 12px;
    border-radius: 4px;
    min-width: 64px;
    text-align: right;
  }
  .form {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-weight: 600;
    font-size: 0.9rem;
  }
  .pair {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }
  .mono {
    font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
    letter-spacing: 0.5px;
  }
  .invalid {
    border-color: var(--danger);
  }
  .error {
    color: var(--danger);
    font-weight: 500;
  }
  .pay {
    padding: 12px 16px;
    font-size: 1rem;
  }
  .tests summary {
    cursor: pointer;
    color: var(--muted);
  }
  .tests ul {
    list-style: none;
    padding: 0;
    margin: 8px 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .tests li {
    display: flex;
    gap: 10px;
    align-items: center;
    flex-wrap: wrap;
  }
  .approved {
    color: var(--success);
  }
  .rejected {
    color: var(--danger);
  }
  @media (max-width: 720px) {
    .card-pay {
      grid-template-columns: 1fr;
    }
    .preview {
      max-width: 320px;
    }
  }
</style>
