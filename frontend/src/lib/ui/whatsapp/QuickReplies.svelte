<script lang="ts">
  import type { MessageOption } from '../../domain';

  interface Props {
    /** The answers the last message from Farmi accepts. */
    options: MessageOption[];
    /** Farmi is waiting for the prescription photo. */
    awaitingPhoto: boolean;
    disabled: boolean;
    onSend: (text: string) => void;
    onAttach: () => void;
  }

  let { options, awaitingPhoto, disabled, onSend, onAttach }: Props = $props();
</script>

{#if options.length > 0 || awaitingPhoto}
  <div class="chips" aria-label="Opciones de respuesta">
    {#if awaitingPhoto}
      <button type="button" class="chip" {disabled} onclick={onAttach}>📎 Adjuntar receta</button>
    {/if}
    {#each options as option (option.value)}
      <button type="button" class="chip" {disabled} onclick={() => onSend(option.value)}>{option.label}</button>
    {/each}
  </div>
{/if}

<style>
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    padding: 8px 10px;
  }
  .chip {
    border: 1px solid rgba(7, 94, 84, 0.3);
    background: rgba(255, 255, 255, 0.85);
    color: var(--wa-dark);
    border-radius: 999px;
    padding: 5px 12px;
    font: inherit;
    font-size: 0.82rem;
    font-weight: 600;
    text-align: left;
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
