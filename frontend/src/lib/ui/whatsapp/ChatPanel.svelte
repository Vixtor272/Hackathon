<script lang="ts">
  import { tick } from 'svelte';
  import type { ConversationState, Message } from '../../domain';
  import ChatBubble from './ChatBubble.svelte';
  import QuickReplies from './QuickReplies.svelte';

  interface Props {
    phone: string;
    messages: Message[];
    conversationState: ConversationState | null;
    connected: boolean;
    busy: boolean;
    onSendText: (text: string) => void;
    onAttach: () => void;
  }

  let { phone, messages, conversationState, connected, busy, onSendText, onAttach }: Props = $props();
  let draft = $state('');
  let body = $state<HTMLDivElement | null>(null);

  // Only the options of Farmi's latest message apply: once the client answers, they go away.
  const options = $derived.by(() => {
    const last = messages.at(-1);
    return last?.direction === 'out' ? (last.options ?? []) : [];
  });
  const awaitingPhoto = $derived(conversationState === 'ASK_PRESCRIPTION' && messages.length > 0);

  $effect(() => {
    void messages.length;
    void options.length;
    void tick().then(() => body?.scrollTo({ top: body.scrollHeight, behavior: 'smooth' }));
  });

  function submit(event: SubmitEvent): void {
    event.preventDefault();
    const text = draft.trim();
    if (text.length === 0 || busy) return;
    onSendText(text);
    draft = '';
  }
</script>

<section class="chat" aria-label="Conversación con Farmi">
  <header class="head">
    <span class="avatar">💊</span>
    <div class="who">
      <strong>Farmi · Farmaenlace</strong>
      <span class="status">{connected ? 'asistente de compras · en línea' : 'sin conexión con el servidor'}</span>
    </div>
    <span class="phone">{phone}</span>
  </header>

  <div class="body" bind:this={body}>
    {#if messages.length === 0}
      <div class="intro">
        <p><strong>Nueva conversación.</strong></p>
        <p>Escribe «hola» para que Farmi te salude y te pida tu cédula.</p>
      </div>
    {/if}
    {#each messages as message (message.id)}
      <ChatBubble {message} />
    {/each}
    {#if busy}
      <div class="typing">Farmi está escribiendo…</div>
    {/if}
  </div>

  <QuickReplies {options} {awaitingPhoto} disabled={busy || !connected} onSend={onSendText} {onAttach} />

  <form class="composer" onsubmit={submit}>
    <button type="button" class="btn attach" onclick={onAttach} disabled={busy || !connected} title="Adjuntar receta">📎 Adjuntar receta</button>
    <input
      class="input"
      placeholder="Escribe un mensaje"
      bind:value={draft}
      disabled={!connected}
      autocomplete="off"
      aria-label="Mensaje"
    />
    <button type="submit" class="btn btn-success send" disabled={busy || !connected || draft.trim().length === 0}>Enviar</button>
  </form>
</section>

<style>
  .chat {
    display: flex;
    flex-direction: column;
    height: min(78vh, 760px);
    min-height: 520px;
    border-radius: var(--radius);
    overflow: hidden;
    border: 1px solid var(--border);
    box-shadow: var(--shadow);
    background: var(--wa-bg);
  }
  .head {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 14px;
    background: var(--wa-dark);
    color: #fff;
  }
  .avatar {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: #fff;
    font-size: 1.2rem;
  }
  .who {
    display: flex;
    flex-direction: column;
    line-height: 1.2;
    flex: 1;
  }
  .status {
    font-size: 0.78rem;
    opacity: 0.85;
  }
  .phone {
    font-size: 0.8rem;
    opacity: 0.8;
  }
  .body {
    flex: 1;
    overflow-y: auto;
    padding: 12px 4px;
    background-image: radial-gradient(rgba(255, 255, 255, 0.35) 1px, transparent 1px);
    background-size: 18px 18px;
  }
  .intro {
    margin: 20px auto;
    max-width: 360px;
    background: #fff8d6;
    border-radius: 10px;
    padding: 12px 14px;
    text-align: center;
    font-size: 0.9rem;
    box-shadow: 0 1px 1px rgba(0, 0, 0, 0.1);
  }
  .intro p {
    margin: 0 0 4px;
  }
  .typing {
    margin: 6px 14px;
    font-size: 0.8rem;
    color: var(--muted);
    font-style: italic;
  }
  .composer {
    display: flex;
    gap: 8px;
    padding: 10px;
    background: #f0f0f0;
    border-top: 1px solid #ddd;
  }
  .composer .input {
    flex: 1;
    border-radius: 999px;
  }
  .attach {
    white-space: nowrap;
  }
  .send {
    border-radius: 999px;
  }
  @media (max-width: 720px) {
    .attach {
      font-size: 0;
      padding: 9px 12px;
    }
    .attach::before {
      content: '📎';
      font-size: 1rem;
    }
  }
</style>
