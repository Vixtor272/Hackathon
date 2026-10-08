import type { Conversation, ConversationState, InboundMessage, Message, SampleMedia, WebhookResult } from '../../domain';
import type { ChatGateway } from '../../application/ports';
import { clone } from './fixtures';

const SAMPLE_MEDIA: SampleMedia[] = [
  { id: 'receta-001', title: 'Receta válida — Dra. Ana Torres', description: 'Paracetamol, Amoxicilina y Loratadina', url: '/recetas/receta-001.svg', expected: 'valid' },
  { id: 'receta-005', title: 'Receta ilegible', description: 'El OCR no puede leer la imagen', url: '/recetas/receta-005.svg', expected: 'illegible' },
];

/** Scripted Farmi: greets, asks for the ID and echoes everything else. Enough for tests. */
export class InMemoryChatGateway implements ChatGateway {
  private readonly conversations = new Map<string, Conversation>();
  private counter = 0;

  async sendInbound(message: InboundMessage): Promise<WebhookResult> {
    const conversation = this.ensure(message.from);
    const inbound: Message =
      message.type === 'text'
        ? this.message('in', 'text', message.text)
        : { ...this.message('in', 'image', '📷 Receta'), mediaId: message.mediaId, mediaUrl: `/recetas/${message.mediaId}.svg` };
    conversation.messages.push(inbound);

    const replies = this.reply(conversation, message);
    conversation.messages.push(...replies);
    return { state: conversation.state, replies: clone(replies) };
  }

  async getConversation(phone: string): Promise<Conversation> {
    return clone(this.ensure(phone));
  }

  async resetConversation(phone: string): Promise<void> {
    this.conversations.delete(phone);
  }

  async listSampleMedia(): Promise<SampleMedia[]> {
    return clone(SAMPLE_MEDIA);
  }

  private ensure(phone: string): Conversation {
    let conversation = this.conversations.get(phone);
    if (!conversation) {
      conversation = { phone, state: 'ASK_ID', orderId: null, messages: [] };
      this.conversations.set(phone, conversation);
    }
    return conversation;
  }

  private reply(conversation: Conversation, message: InboundMessage): Message[] {
    if (conversation.state === 'ASK_ID') {
      if (message.type === 'text' && /^\d{10}$/.test(message.text)) {
        conversation.state = 'ASK_PRESCRIPTION';
        return [this.message('out', 'text', 'Gracias. Envíame una foto de tu receta.')];
      }
      return [this.message('out', 'text', '¡Hola! Soy Farmi. Antes de empezar, ¿me indicas tu número de cédula?')];
    }
    if (conversation.state === 'ASK_PRESCRIPTION' && message.type === 'image') {
      conversation.state = 'ASK_ZONE';
      return [this.message('out', 'text', 'Receta validada. ¿En qué ciudad o zona deseas comprar?')];
    }
    return [this.message('out', 'text', `Recibí: ${message.type === 'text' ? message.text : message.mediaId}`)];
  }

  private message(direction: Message['direction'], type: Message['type'], text: string): Message {
    this.counter += 1;
    return { id: `msg_${this.counter}`, direction, type, text, at: new Date(0).toISOString() };
  }

  /** Test helper: force a state. */
  setState(phone: string, state: ConversationState): void {
    this.ensure(phone).state = state;
  }
}
