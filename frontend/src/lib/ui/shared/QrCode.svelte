<script lang="ts">
  import { qrMatrix } from '../../qr';

  interface Props {
    value: string;
    size?: number;
    label?: string;
  }

  let { value, size = 220, label = 'Código QR' }: Props = $props();

  const QUIET = 4; // modules of white border scanners need around the code

  const matrix = $derived(qrMatrix(value));
  const span = $derived(matrix.length + QUIET * 2);
  const path = $derived(
    matrix
      .flatMap((row, r) => row.map((dark, c) => (dark ? `M${c + QUIET},${r + QUIET}h1v1h-1z` : '')))
      .join(''),
  );
</script>

<svg
  class="qr"
  width={size}
  height={size}
  viewBox={`0 0 ${span} ${span}`}
  shape-rendering="crispEdges"
  role="img"
  aria-label={label}
>
  <rect width={span} height={span} fill="#fff" />
  <path d={path} fill="#111" />
</svg>

<style>
  .qr {
    display: block;
    border-radius: 10px;
    max-width: 100%;
    height: auto;
  }
</style>
