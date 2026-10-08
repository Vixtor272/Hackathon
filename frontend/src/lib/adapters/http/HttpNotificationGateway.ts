import type { Notification } from '../../domain';
import type { NotificationFilter, NotificationGateway } from '../../application/ports';
import type { HttpClient } from './httpClient';

export class HttpNotificationGateway implements NotificationGateway {
  constructor(private readonly http: HttpClient) {}

  async listNotifications(filter: NotificationFilter = {}): Promise<Notification[]> {
    const { notifications } = await this.http.get<{ notifications: Notification[] }>('/notifications', {
      channel: filter.channel,
      orderId: filter.orderId,
    });
    return notifications;
  }
}
