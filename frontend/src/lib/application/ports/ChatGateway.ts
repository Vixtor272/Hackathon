import type { Conversation, InboundMessage, SampleMedia, WebhookResult } from '../../domain';

/** Driven port: the (simulated) WhatsApp channel between the client and Farmi. */
export interface ChatGateway {
  sendInbound(message: InboundMessage): Promise<WebhookResult>;
  getConversation(phone: string): Promise<Conversation>;
  resetConversation(phone: string): Promise<void>;
  listSampleMedia(): Promise<SampleMedia[]>;
}
