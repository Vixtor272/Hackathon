import type { Notification, NotificationChannel } from '../../domain';

export interface NotificationFilter {
  channel?: NotificationChannel;
  orderId?: string;
}

/** Driven port: the outbox of every simulated notification channel. */
export interface NotificationGateway {
  listNotifications(filter?: NotificationFilter): Promise<Notification[]>;
}
