import type { Bank, Payment, PaymentConfirmation, PaymentMethod, PaymentOutcome, PaymentWithOrder } from '../../domain';
import type { PaymentGateway } from '../../application/ports';
import type { HttpClient } from './httpClient';

export class HttpPaymentGateway implements PaymentGateway {
  constructor(private readonly http: HttpClient) {}

  async createPayment(orderId: string, method: PaymentMethod): Promise<Payment> {
    const { payment } = await this.http.post<{ payment: Payment }>('/payments', { orderId, method });
    return payment;
  }

  getPayment(paymentId: string): Promise<PaymentWithOrder> {
    return this.http.get<PaymentWithOrder>(`/payments/${encodeURIComponent(paymentId)}`);
  }

  async listBanks(): Promise<Bank[]> {
    const { banks } = await this.http.get<{ banks: Bank[] }>('/payments/banks');
    return banks;
  }

  confirmPayment(paymentId: string, outcome: PaymentOutcome, bankId?: string): Promise<PaymentConfirmation> {
    return this.http.post<PaymentConfirmation>(`/payments/${encodeURIComponent(paymentId)}/confirm`, {
      outcome,
      ...(bankId ? { bankId } : {}),
    });
  }
}
