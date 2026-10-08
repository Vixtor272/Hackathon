<script lang="ts">
  import { QUICK_REPLIES } from './contacts';

  interface Props {
    disabled: boolean;
    onSend: (text: string) => void;
  }

  let { disabled, onSend }: Props = $props();
</script>

<div class="chips" aria-label="Respuestas rápidas">
  {#each QUICK_REPLIES as reply (reply)}
    <button type="button" class="chip" {disabled} onclick={() => onSend(reply)}>{reply}</button>
  {/each}
</div>

<style>
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    padding: 8px 10px 0;
  }
  .chip {
    border: 1px solid rgba(7, 94, 84, 0.3);
    background: rgba(255, 255, 255, 0.85);
    color: var(--wa-dark);
    border-radius: 999px;
    padding: 4px 12px;
    font: inherit;
    font-size: 0.82rem;
    font-weight: 600;
    cursor: pointer;
  }
  .chip:hover:not(:disabled) {
    background: #fff;
  }
  .chip:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
