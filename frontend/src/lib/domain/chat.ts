export type MessageDirection = 'in' | 'out';
export type MessageType = 'text' | 'image' | 'link';

/** A quick reply offered with a message: `value` is the text sent when it is tapped. */
export interface MessageOption {
  label: string;
  value: string;
}

export interface Message {
  id: string;
  direction: MessageDirection;
  type: MessageType;
  text: string;
  mediaId?: string;
  mediaUrl?: string;
  link?: string;
  /** Answers Farmi expects to this message; only on the last reply of a turn. */
  options?: MessageOption[];
  at: string;
}

export type ConversationState =
  | 'ASK_ID'
  | 'ASK_NAME'
  | 'ASK_PRESCRIPTION'
  | 'ASK_ZONE'
  | 'ASK_MODE'
  | 'ASK_PICKUP_OPTION'
  | 'ASK_ADDRESS'
  | 'ASK_BRAND'
  | 'CONFIRM_CART'
  | 'AWAIT_PAYMENT'
  | 'COMPLETED';

export interface Conversation {
  phone: string;
  state: ConversationState;
  orderId: string | null;
  messages: Message[];
}

export type InboundMessage =
  | { from: string; type: 'text'; text: string }
  | { from: string; type: 'image'; mediaId: string };

export interface WebhookResult {
  state: ConversationState;
  replies: Message[];
}

export type SampleExpectation = 'valid' | 'doctor_inactive' | 'doctor_unknown' | 'incomplete' | 'illegible';

/** A sample prescription image the simulator can attach instead of a real photo. */
export interface SampleMedia {
  id: string;
  title: string;
  description: string;
  url: string;
  expected: SampleExpectation;
}

export const CONVERSATION_STATE_LABELS: Record<ConversationState, string> = {
  ASK_ID: 'Pidiendo cédula',
  ASK_NAME: 'Pidiendo nombre',
  ASK_PRESCRIPTION: 'Esperando receta',
  ASK_ZONE: 'Eligiendo zona',
  ASK_MODE: 'Retiro o domicilio',
  ASK_PICKUP_OPTION: 'Eligiendo farmacia',
  ASK_ADDRESS: 'Pidiendo dirección',
  ASK_BRAND: 'Eligiendo marcas',
  CONFIRM_CART: 'Confirmando carrito',
  AWAIT_PAYMENT: 'Esperando pago',
  COMPLETED: 'Compra completada',
};
