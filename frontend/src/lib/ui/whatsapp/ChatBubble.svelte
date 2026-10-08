<script lang="ts">
  import type { Message } from '../../domain';
  import { formatTime } from '../../format';
  import { openLink } from '../../router';

  interface Props {
    message: Message;
  }

  let { message }: Props = $props();

  /** Backend "in" = sent by the client → right-hand green bubble, like WhatsApp. */
  const mine = $derived(message.direction === 'in');
</script>

<div class="line" class:mine>
  <div class="bubble" class:mine>
    {#if message.type === 'image' && message.mediaUrl}
      <img class="media" src={message.mediaUrl} alt="Receta enviada" loading="lazy" />
    {/if}
    {#if message.text}
      <p class="text">{message.text}</p>
    {/if}
    {#if message.type === 'link' && message.link}
      <button type="button" class="btn btn-primary btn-sm link" onclick={() => openLink(message.link ?? '')}>
        Abrir página de pago →
      </button>
    {/if}
    <span class="time">{formatTime(message.at)}{#if mine} ✓✓{/if}</span>
  </div>
</div>

<style>
  .line {
    display: flex;
    justify-content: flex-start;
    padding: 2px 10px;
  }
  .line.mine {
    justify-content: flex-end;
  }
  .bubble {
    max-width: min(78%, 480px);
    background: #fff;
    border-radius: 10px;
    padding: 7px 10px 18px;
    position: relative;
    box-shadow: 0 1px 1px rgba(0, 0, 0, 0.12);
    font-size: 0.95rem;
  }
  .bubble.mine {
    background: var(--wa-bubble-out);
  }
  .text {
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
  }
  .media {
    display: block;
    width: 220px;
    max-width: 100%;
    border-radius: 8px;
    margin-bottom: 6px;
    border: 1px solid rgba(0, 0, 0, 0.08);
    background: #fff;
  }
  .link {
    margin-top: 8px;
  }
  .time {
    position: absolute;
    right: 8px;
    bottom: 3px;
    font-size: 0.68rem;
    color: rgba(0, 0, 0, 0.45);
  }
</style>
