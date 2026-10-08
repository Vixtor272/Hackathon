<script lang="ts">
  import type { Message } from '../../domain';
  import { formatTime } from '../../format';
  import { openLink } from '../../router';
  import { parseMarkup } from './markup';

  interface Props {
    message: Message;
  }

  let { message }: Props = $props();

  /** Backend "in" = sent by the client → right-hand green bubble, like WhatsApp. */
  const mine = $derived(message.direction === 'in');
  const lines = $derived(parseMarkup(message.text));
</script>

<div class="line" class:mine>
  <div class="bubble" class:mine>
    {#if message.type === 'image' && message.mediaUrl}
      <img class="media" src={message.mediaUrl} alt="Receta enviada" loading="lazy" />
    {/if}
    {#if message.text}
      <div class="text">
        {#each lines as line, index (index)}
          <div class="row" class:title={line.title}>
            {#each line.spans as span, at (at)}{#if span.bold}<strong>{span.text}</strong>{:else}{span.text}{/if}{/each}
          </div>
        {/each}
      </div>
    {/if}
    {#if message.type === 'link' && message.link}
      <button type="button" class="btn btn-primary btn-sm link" onclick={() => openLink(message.link ?? '')}>
        Abrir página de pago →
      </button>
    {/if}
  </div>
  <span class="time">{formatTime(message.at)}{#if mine} ✓✓{/if}</span>
</div>

<style>
  .line {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    padding: 3px 10px;
  }
  .line.mine {
    align-items: flex-end;
  }
  .bubble {
    max-width: min(78%, 480px);
    background: #fff;
    border-radius: 10px;
    padding: 7px 10px;
    box-shadow: 0 1px 1px rgba(0, 0, 0, 0.12);
    font-size: 0.95rem;
  }
  .bubble.mine {
    background: var(--wa-bubble-out);
  }
  .text {
    white-space: pre-wrap;
    word-break: break-word;
  }
  .row {
    min-height: 1.35em;
  }
  .row.title {
    font-size: 1.02rem;
    margin-bottom: 3px;
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
    margin: 2px 4px 0;
    font-size: 0.68rem;
    color: rgba(0, 0, 0, 0.5);
  }
</style>
