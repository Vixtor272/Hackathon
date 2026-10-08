import type { Bank, Payment, PaymentConfirmation, PaymentMethod, PaymentOutcome, PaymentWithOrder } from '../../domain';

/** Driven port: payment intents and the card / DeUna simulators. */
export interface PaymentGateway {
  createPayment(orderId: string, method: PaymentMethod): Promise<Payment>;
  getPayment(paymentId: string): Promise<PaymentWithOrder>;
  listBanks(): Promise<Bank[]>;
  confirmPayment(paymentId: string, outcome: PaymentOutcome, bankId?: string): Promise<PaymentConfirmation>;
}
