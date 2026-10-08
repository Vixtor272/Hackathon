import { ApiError, type Bank, type Payment, type PaymentConfirmation, type PaymentOutcome, type PaymentWithOrder } from '../../domain';
import type { PaymentGateway } from '../ports';

/** Card and DeUna simulators: create an intent, then confirm or reject it. */
export class PaymentFlow {
  constructor(private readonly payments: PaymentGateway) {}

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
