import { ApiError, type Conversation, type SampleMedia, type WebhookResult } from '../../domain';
import type { ChatGateway } from '../ports';

/** Drives one WhatsApp conversation between a phone number and Farmi. */
export class ChatSession {
  constructor(private readonly chat: ChatGateway) {}

  loadTranscript(phone: string): Promise<Conversation> {
    return this.chat.getConversation(normalizePhone(phone));
  }

  sendText(phone: string, text: string): Promise<WebhookResult> {
    const trimmed = text.trim();
    if (trimmed.length === 0) {
      return Promise.reject(new ApiError('VALIDATION', 'Escribe un mensaje antes de enviar'));
    }
    return this.chat.sendInbound({ from: normalizePhone(phone), type: 'text', text: trimmed });
  }

  sendImage(phone: string, mediaId: string): Promise<WebhookResult> {
    if (mediaId.trim().length === 0) {
      return Promise.reject(new ApiError('VALIDATION', 'Selecciona una receta para enviar'));
    }
    return this.chat.sendInbound({ from: normalizePhone(phone), type: 'image', mediaId });
  }

  reset(phone: string): Promise<void> {
    return this.chat.resetConversation(normalizePhone(phone));
  }

  listSampleMedia(): Promise<SampleMedia[]> {
    return this.chat.listSampleMedia();
  }
}

/** Keeps only the leading "+" and digits so "+593 99 111 1111" and "+593991111111" are the same chat. */
export function normalizePhone(phone: string): string {
  const compact = phone.replace(/[^\d+]/g, '');
  const digits = compact.replace(/\+/g, '');
  return compact.startsWith('+') ? `+${digits}` : digits;
}
