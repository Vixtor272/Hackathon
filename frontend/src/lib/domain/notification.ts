export type NotificationChannel = 'whatsapp' | 'pharmacy' | 'courier';

export interface Notification {
  id: string;
  channel: NotificationChannel;
  recipient: string;
  title: string;
  body: string;
  orderId: string;
  at: string;
}

export const NOTIFICATION_CHANNEL_LABELS: Record<NotificationChannel, string> = {
  whatsapp: 'WhatsApp',
  pharmacy: 'Farmacia',
  courier: 'Repartidor',
};
