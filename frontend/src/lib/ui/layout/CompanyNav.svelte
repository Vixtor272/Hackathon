<script lang="ts">
  const scope = 'Caja · Reparto · Notificaciones';
</script>

<header class="topnav">
  <span class="brand">
    <span class="logo">🏪</span>
    <span>Farmaenlace <span class="muted">· Portal empresa</span></span>
  </span>
  <span class="scope">{scope}</span>
  <span class="tag">Uso interno · datos simulados</span>
</header>

<style>
  .topnav {
    display: flex;
    align-items: center;
    gap: 24px;
    padding: 10px 20px;
    background: #1e293b;
    color: #f8fafc;
    position: sticky;
    top: 0;
    z-index: 10;
  }
  .brand {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-weight: 700;
    font-size: 1.05rem;
  }
  .brand .muted {
    color: #cbd5e1;
  }
  .logo {
    display: inline-grid;
    place-items: center;
    width: 32px;
    height: 32px;
    border-radius: 9px;
    background: #334155;
  }
  .scope {
    flex: 1;
    color: #cbd5e1;
    font-weight: 600;
  }
  .tag {
    font-size: 0.75rem;
    color: #cbd5e1;
    border: 1px solid #475569;
    border-radius: 999px;
    padding: 3px 10px;
  }
  @media (max-width: 720px) {
    .topnav {
      gap: 12px;
      padding: 8px 12px;
    }
    .scope,
    .tag {
      display: none;
    }
  }
</style>
