import type { Notification } from '../../domain';
import type { NotificationFilter, NotificationGateway } from '../../application/ports';
import { clone } from './fixtures';

export class InMemoryNotificationGateway implements NotificationGateway {
  constructor(private readonly notifications: Notification[] = []) {}

  async listNotifications(filter: NotificationFilter = {}): Promise<Notification[]> {
    return clone(
      this.notifications.filter(
        (notification) =>
          (!filter.channel || notification.channel === filter.channel) &&
          (!filter.orderId || notification.orderId === filter.orderId),
      ),
    );
  }
}
