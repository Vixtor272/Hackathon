<script lang="ts">
  import { formatCountdown } from '../../format';

  interface Props {
    /** Seconds left as reported by the server; the component ticks locally in between polls. */
    secondsLeft: number;
    active: boolean;
  }

  let { secondsLeft, active }: Props = $props();
  let remaining = $state(0);

  $effect(() => {
    remaining = secondsLeft;
  });

  $effect(() => {
    if (!active) return;
    const id = setInterval(() => {
      remaining = Math.max(0, remaining - 1);
    }, 1000);
    return () => clearInterval(id);
  });

  const urgent = $derived(active && remaining > 0 && remaining < 120);
</script>

<span class="countdown" class:urgent class:done={!active || remaining === 0}>
  {formatCountdown(remaining)}
</span>

<style>
  .countdown {
    font-variant-numeric: tabular-nums;
    font-weight: 700;
    font-size: 1.1rem;
    color: var(--primary-strong);
  }
  .urgent {
    color: var(--warning);
  }
  .done {
    color: var(--danger);
  }
</style>
