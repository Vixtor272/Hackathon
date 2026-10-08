<script lang="ts">
  import type { SampleMedia } from '../../domain';
  import StatusPill from '../shared/StatusPill.svelte';
  import { EXPECTATION_LABELS } from '../shared/tones';

  interface Props {
    media: SampleMedia[];
    loading: boolean;
    error: string | null;
    onPick: (mediaId: string) => void;
    onClose: () => void;
  }

  let { media, loading, error, onPick, onClose }: Props = $props();
</script>

<button type="button" class="backdrop" aria-label="Cerrar selector" onclick={onClose}></button>
<div class="modal" role="dialog" aria-modal="true" aria-labelledby="picker-title">
  <div class="row between">
    <h2 id="picker-title">Adjuntar receta</h2>
    <button type="button" class="btn btn-ghost btn-sm" onclick={onClose} aria-label="Cerrar">✕</button>
  </div>
  <p class="muted small">Estas imágenes reemplazan la foto que el cliente tomaría con su teléfono.</p>

  {#if loading}
    <p class="loading">Cargando recetas de ejemplo…</p>
  {:else if error}
    <p class="banner banner-danger">{error}</p>
  {:else if media.length === 0}
    <p class="empty">No hay recetas de ejemplo disponibles.</p>
  {:else}
    <div class="grid">
      {#each media as item (item.id)}
        <button type="button" class="option" onclick={() => onPick(item.id)}>
          <img src={item.url} alt={item.title} loading="lazy" />
          <span class="body">
            <strong>{item.title}</strong>
            <span class="muted small">{item.description}</span>
            <StatusPill label={EXPECTATION_LABELS[item.expected].label} tone={EXPECTATION_LABELS[item.expected].tone} />
          </span>
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgba(16, 24, 40, 0.45);
    border: 0;
    z-index: 20;
    cursor: default;
  }
  .modal {
    position: fixed;
    z-index: 21;
    left: 50%;
    top: 50%;
    transform: translate(-50%, -50%);
    width: min(860px, calc(100vw - 32px));
    max-height: calc(100vh - 48px);
    overflow: auto;
    background: var(--surface);
    border-radius: 14px;
    padding: 18px 20px;
    box-shadow: 0 20px 60px rgba(16, 24, 40, 0.3);
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
    gap: 12px;
  }
  .option {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 10px;
    border: 1px solid var(--border);
    border-radius: 12px;
    background: var(--surface);
    cursor: pointer;
    text-align: left;
    font: inherit;
    color: inherit;
  }
  .option:hover {
    border-color: var(--primary);
    background: var(--primary-soft);
  }
  .option img {
    width: 100%;
    aspect-ratio: 3 / 4;
    object-fit: cover;
    object-position: top;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: #fff;
  }
  .body {
    display: flex;
    flex-direction: column;
    gap: 4px;
    align-items: flex-start;
  }
</style>
