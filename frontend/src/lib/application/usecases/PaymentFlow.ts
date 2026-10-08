import {
  ApiError,
  hasErrors,
  randomReference,
  simulatedOutcome,
  validateCard,
  type Bank,
  type CardForm,
  type Payment,
  type PaymentConfirmation,
  type PaymentOutcome,
  type PaymentWithOrder,
} from '../../domain';

/** A DeUna request shown as a QR on a computer, to be scanned with the phone. */
export interface DeUnaQr {
  payment: Payment;
  link: string;
  reference: string;
  qrPayload: string;
}
import type { PaymentGateway } from '../ports';

/** Card and DeUna simulators: create an intent, then confirm or reject it. */
export class PaymentFlow {
  constructor(
    private readonly payments: PaymentGateway,
    private readonly newReference: () => string = randomReference,
  ) {}

  startCard(orderId: string): Promise<Payment> {
    return this.payments.createPayment(orderId, 'card');
  }

  async startDeUna(orderId: string): Promise<{ payment: Payment; link: string }> {
    const payment = await this.payments.createPayment(orderId, 'deuna');
    if (!payment.link) {
      throw new ApiError('INTERNAL', 'DeUna no devolvió un enlace de pago');
    }
    return { payment, link: payment.link };
  }

  /** Card simulator: one click creates the intent and resolves it with the chosen outcome. */
  async payWithCard(orderId: string, outcome: PaymentOutcome): Promise<PaymentConfirmation> {
    const payment = await this.startCard(orderId);
    return this.confirm(payment.id, outcome);
  }

  /**
   * Card form: the details are validated here and never leave the browser (it
   * is a simulator); the sandbox card number decides approval or rejection.
   */
  payWithCardDetails(orderId: string, form: CardForm, now: Date = new Date()): Promise<PaymentConfirmation> {
    if (hasErrors(validateCard(form, now))) {
      return Promise.reject(new ApiError('VALIDATION', 'Revisa los datos de la tarjeta'));
    }
    return this.payWithCard(orderId, simulatedOutcome(form.number));
  }

  /** DeUna from a computer: the same payment intent, presented as a QR with a fresh random reference. */
  async startDeUnaQr(orderId: string): Promise<DeUnaQr> {
    const { payment, link } = await this.startDeUna(orderId);
    const reference = this.newReference();
    const url = new URL(link);
    url.searchParams.set('ref', reference);
    return { payment, link, reference, qrPayload: url.toString() };
  }

  /** The QR is paid from the client's DeUna app, which already knows their bank account. */
  confirmDeUnaQr(paymentId: string, outcome: PaymentOutcome): Promise<PaymentConfirmation> {
    return this.confirm(paymentId, outcome);
  }

  loadPayment(paymentId: string): Promise<PaymentWithOrder> {
    return this.payments.getPayment(paymentId);
  }

  listBanks(): Promise<Bank[]> {
    return this.payments.listBanks();
  }

  confirm(paymentId: string, outcome: PaymentOutcome, bankId?: string): Promise<PaymentConfirmation> {
    return this.payments.confirmPayment(paymentId, outcome, bankId);
  }

  /** DeUna requires a bank before confirming; rejection does not. */
  confirmDeUna(paymentId: string, outcome: PaymentOutcome, bankId: string | null): Promise<PaymentConfirmation> {
    if (outcome === 'approved' && !bankId) {
      return Promise.reject(new ApiError('VALIDATION', 'Elige un banco para confirmar el pago'));
    }
    return this.confirm(paymentId, outcome, bankId ?? undefined);
  }
}
