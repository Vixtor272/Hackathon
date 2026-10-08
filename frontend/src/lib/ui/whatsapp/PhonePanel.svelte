<script lang="ts">
  import { CONVERSATION_STATE_LABELS, type ConversationState } from '../../domain';
  import StatusPill from '../shared/StatusPill.svelte';
  import { DEMO_CONTACTS } from './contacts';

  interface Props {
    phone: string;
    conversationState: ConversationState | null;
    orderId: string | null;
    busy: boolean;
    onSelect: (phone: string) => void;
    onReset: () => void;
  }

  let { phone, conversationState, orderId, busy, onSelect, onReset }: Props = $props();
  let customPhone = $state('');

  const isDemo = $derived(DEMO_CONTACTS.some((contact) => contact.phone === phone));

  function useCustom(event: SubmitEvent): void {
    event.preventDefault();
    const value = customPhone.trim();
    if (value.length > 0) onSelect(value);
  }
</script>

<aside class="card panel">
  <h2>Teléfono del cliente</h2>
  <p class="muted small">Elige quién chatea con Farmi. Cada número tiene su propia conversación.</p>

  <ul class="contacts">
    {#each DEMO_CONTACTS as contact (contact.phone)}
      <li>
        <button
          type="button"
          class="contact"
          class:selected={contact.phone === phone}
          onclick={() => onSelect(contact.phone)}
        >
          <span class="avatar">{contact.name.charAt(0)}</span>
          <span class="info">
            <strong>{contact.name}</strong>
            <span class="muted small">{contact.phone}</span>
            <span class="muted small">Cédula {contact.idNumber}</span>
          </span>
        </button>
      </li>
    {/each}
  </ul>

  <form class="custom" onsubmit={useCustom}>
    <label class="small muted" for="custom-phone">Otro número</label>
    <div class="row">
      <input id="custom-phone" class="input" placeholder="+5939…" bind:value={customPhone} />
      <button type="submit" class="btn btn-sm">Usar</button>
    </div>
  </form>

  {#if !isDemo}
    <p class="small muted">Chateando como <strong>{phone}</strong> (cliente nuevo: Farmi pedirá cédula y nombre).</p>
  {/if}

  <div class="state">
    <span class="small muted">Estado de la conversación</span>
    {#if conversationState}
      <StatusPill label={CONVERSATION_STATE_LABELS[conversationState]} tone={conversationState === 'COMPLETED' ? 'success' : conversationState === 'AWAIT_PAYMENT' ? 'warning' : 'info'} />
    {:else}
      <StatusPill label="Sin conexión" tone="neutral" />
    {/if}
    {#if orderId}
      <a class="small" href={`/checkout/${encodeURIComponent(orderId)}`}>Ver pedido actual →</a>
    {/if}
  </div>

  <button type="button" class="btn btn-danger" disabled={busy} onclick={onReset}>Reiniciar conversación</button>

  <details class="help">
    <summary class="small">¿Cómo probar el flujo?</summary>
    <ol class="small muted">
      <li>Escribe «hola» y responde con la cédula.</li>
      <li>Adjunta una receta con 📎 (prueba la válida y las inválidas).</li>
      <li>Elige zona, retiro o domicilio, y marcas.</li>
      <li>Confirma y abre el enlace de pago.</li>
    </ol>
  </details>
</aside>

<style>
  .panel {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  .contacts {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .contact {
    width: 100%;
    display: flex;
    gap: 10px;
    align-items: center;
    padding: 8px 10px;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--surface);
    cursor: pointer;
    text-align: left;
    font: inherit;
    color: inherit;
  }
  .contact:hover {
    background: var(--surface-muted);
  }
  .contact.selected {
    border-color: var(--primary);
    background: var(--primary-soft);
  }
  .avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: var(--wa-dark);
    color: #fff;
    font-weight: 700;
    flex-shrink: 0;
  }
  .info {
    display: flex;
    flex-direction: column;
    line-height: 1.25;
  }
  .custom {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .custom .row {
    flex-wrap: nowrap;
  }
  .state {
    display: flex;
    flex-direction: column;
    gap: 6px;
    align-items: flex-start;
  }
  .help summary {
    cursor: pointer;
    color: var(--primary);
  }
  .help ol {
    padding-left: 18px;
    margin: 6px 0 0;
  }
</style>
