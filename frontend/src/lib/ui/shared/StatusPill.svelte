<script lang="ts">
  import type { Tone } from './tones';

  interface Props {
    label: string;
    tone?: Tone;
  }

  let { label, tone = 'neutral' }: Props = $props();
</script>

<span class="pill pill-{tone}">{label}</span>
