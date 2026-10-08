<script lang="ts">
  import { ROUTES } from '../router';

  interface Props {
    path: string;
  }

  let { path }: Props = $props();
</script>

<section class="card">
  <h1>Página no encontrada</h1>
  <p class="muted">No existe nada en <code>{path}</code>.</p>
  <a class="btn btn-primary" href={ROUTES.whatsapp}>Ir al simulador de WhatsApp</a>
</section>
