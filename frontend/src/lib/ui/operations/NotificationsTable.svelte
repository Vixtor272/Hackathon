<script lang="ts">
  import { NOTIFICATION_CHANNEL_LABELS, type Notification } from '../../domain';
  import { formatDateTime } from '../../format';
  import StatusPill from '../shared/StatusPill.svelte';

  interface Props {
    notifications: Notification[];
  }

  let { notifications }: Props = $props();

  const ordered = $derived([...notifications].sort((a, b) => b.at.localeCompare(a.at)));
</script>

{#if ordered.length === 0}
  <p class="empty">Aún no se ha enviado ninguna notificación.</p>
{:else}
  <div class="scroll">
    <table class="table">
      <thead>
        <tr>
          <th>Hora</th>
          <th>Canal</th>
          <th>Destinatario</th>
          <th>Mensaje</th>
        </tr>
      </thead>
      <tbody>
        {#each ordered as notification (notification.id)}
          <tr>
            <td class="small nowrap">{formatDateTime(notification.at)}</td>
            <td>
              <StatusPill
                label={NOTIFICATION_CHANNEL_LABELS[notification.channel]}
                tone={notification.channel === 'whatsapp' ? 'success' : notification.channel === 'pharmacy' ? 'info' : 'warning'}
              />
            </td>
            <td class="small">{notification.recipient}</td>
            <td>
              <strong>{notification.title}</strong>
              <div class="small muted body">{notification.body}</div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
{/if}

<style>
  .scroll {
    overflow-x: auto;
  }
  .nowrap {
    white-space: nowrap;
  }
  .body {
    white-space: pre-wrap;
  }
</style>
