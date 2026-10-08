import type { Conversation, InboundMessage, SampleMedia, WebhookResult } from '../../domain';
import type { ChatGateway } from '../../application/ports';
import type { HttpClient } from './httpClient';

export class HttpChatGateway implements ChatGateway {
  constructor(private readonly http: HttpClient) {}

  sendInbound(message: InboundMessage): Promise<WebhookResult> {
    return this.http.post<WebhookResult>('/whatsapp/webhook', message);
  }

  getConversation(phone: string): Promise<Conversation> {
    return this.http.get<Conversation>(`/whatsapp/conversations/${encodeURIComponent(phone)}/messages`);
  }

  resetConversation(phone: string): Promise<void> {
    return this.http.delete(`/whatsapp/conversations/${encodeURIComponent(phone)}`);
  }

  async listSampleMedia(): Promise<SampleMedia[]> {
    const response = await this.http.get<{ media: SampleMedia[] }>('/whatsapp/media');
    return response.media;
  }
}
