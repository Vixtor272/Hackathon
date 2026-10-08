<script lang="ts">
  interface Props {
    message: string | null;
    tone?: 'danger' | 'warning' | 'info' | 'success';
    onclose?: () => void;
  }

  let { message, tone = 'danger', onclose }: Props = $props();
</script>

{#if message}
  <div class="banner banner-{tone}" role="alert">
    <span class="text">{message}</span>
    {#if onclose}
      <button type="button" class="btn btn-ghost btn-sm close" onclick={onclose} aria-label="Cerrar aviso">✕</button>
    {/if}
  </div>
{/if}

<style>
  .text {
    flex: 1;
  }
  .close {
    margin: -4px -8px -4px 0;
    color: inherit;
  }
</style>
