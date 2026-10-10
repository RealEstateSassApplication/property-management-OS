import { apiRequestHeaders } from "./request-identity";

export type RentAdjustment = {
  id: string;
  organizationId: string;
  obligationId: string;
  leaseReference: string;
  tenantName: string;
  adjustmentType: "charge" | "late_fee" | "credit" | "writeoff" | "payment_reversal";
  amountMinor: number;
  currency: string;
  reason: string;
  createdAt: string;
};

export type PaymentReversal = {
  id: string;
  organizationId: string;
  paymentId: string;
  tenantName: string;
  amountMinor: number;
  currency: string;
  referenceCode?: string;
  reason: string;
  reversedAt: string;
};

export type DepositAccount = {
  id: string;
  organizationId: string;
  leaseId: string;
  leaseReference: string;
  tenantName: string;
  propertyName: string;
  unitLabel: string;
  requiredAmountMinor: number;
  heldAmountMinor: number;
  currency: string;
  createdAt: string;
};

export type DepositTransaction = {
  id: string;
  organizationId: string;
  depositAccountId: string;
  transactionType: "received" | "deduction" | "refund" | "adjustment_increase" | "adjustment_decrease";
  amountMinor: number;
  occurredOn: string;
  note: string;
  createdAt: string;
};

export type PropertyExpense = {
  id: string;
  organizationId: string;
  propertyId: string;
  propertyName: string;
  vendorId?: string;
  vendorName?: string;
  workOrderId?: string;
  category: string;
  amountMinor: number;
  currency: string;
  incurredOn: string;
  note: string;
  referenceCode?: string;
  createdAt: string;
  reversed: boolean;
  reversalReason?: string;
  reversedAt?: string;
};

export type StatementLine = {
  date: string;
  lineType: "income" | "expense";
  propertyId: string;
  propertyName: string;
  description: string;
  currency: string;
  grossMinor: number;
  ownershipBps: number;
  ownerMinor: number;
};

export type OwnerStatement = {
  ownerId: string;
  ownerName: string;
  from: string;
  to: string;
  summaries: { currency: string; incomeMinor: number; expenseMinor: number; netOwnerAmountMinor: number }[];
  lines: StatementLine[];
};

const apiBaseUrl = process.env.API_BASE_URL ?? "http://localhost:8080";

export class AccountingConfigurationError extends Error {}

async function headers(): Promise<Record<string, string>> {
  try {
    const identity = await apiRequestHeaders();
    return identity;
  } catch (error) {
    throw new AccountingConfigurationError(error instanceof Error ? error.message : "Property OS authentication is not configured.");
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers: { ...await headers(), ...(init?.headers ?? {}) },
    cache: "no-store",
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? `Property OS API returned ${response.status}`);
  }
  return response.json() as Promise<T>;
}

export async function listRentAdjustments(): Promise<RentAdjustment[]> {
  return (await request<{ data: RentAdjustment[] }>("/api/v1/accounting/rent-adjustments")).data;
}
export async function createRentAdjustment(input: { obligationId: string; adjustmentType: string; amountMinor: number; reason: string }): Promise<RentAdjustment> {
  return (await request<{ data: RentAdjustment }>("/api/v1/accounting/rent-adjustments", { method: "POST", body: JSON.stringify(input) })).data;
}
export async function listPaymentReversals(): Promise<PaymentReversal[]> {
  return (await request<{ data: PaymentReversal[] }>("/api/v1/accounting/payment-reversals")).data;
}
export async function reversePayment(paymentId: string, reason: string): Promise<PaymentReversal> {
  return (await request<{ data: PaymentReversal }>(`/api/v1/accounting/payments/${paymentId}/reverse`, { method: "POST", body: JSON.stringify({ reason }) })).data;
}
export async function listDepositAccounts(): Promise<DepositAccount[]> {
  return (await request<{ data: DepositAccount[] }>("/api/v1/accounting/security-deposits")).data;
}
export async function createDepositAccount(leaseId: string): Promise<DepositAccount> {
  return (await request<{ data: DepositAccount }>("/api/v1/accounting/security-deposits", { method: "POST", body: JSON.stringify({ leaseId }) })).data;
}
export async function listDepositTransactions(): Promise<DepositTransaction[]> {
  return (await request<{ data: DepositTransaction[] }>("/api/v1/accounting/security-deposit-transactions")).data;
}
export async function createDepositTransaction(input: { depositAccountId: string; transactionType: string; amountMinor: number; occurredOn: string; note: string }): Promise<DepositTransaction> {
  return (await request<{ data: DepositTransaction }>("/api/v1/accounting/security-deposit-transactions", { method: "POST", body: JSON.stringify(input) })).data;
}
export async function listExpenses(): Promise<PropertyExpense[]> {
  return (await request<{ data: PropertyExpense[] }>("/api/v1/accounting/expenses")).data;
}
export async function createExpense(input: { propertyId: string; vendorId?: string; workOrderId?: string; category: string; amountMinor: number; currency: string; incurredOn: string; note: string; referenceCode?: string }): Promise<PropertyExpense> {
  return (await request<{ data: PropertyExpense }>("/api/v1/accounting/expenses", { method: "POST", body: JSON.stringify(input) })).data;
}
export async function reverseExpense(expenseId: string, reason: string): Promise<PropertyExpense> {
  return (await request<{ data: PropertyExpense }>(`/api/v1/accounting/expenses/${expenseId}/reverse`, { method: "POST", body: JSON.stringify({ reason }) })).data;
}
export async function getOwnerStatement(ownerId: string, from: string, to: string): Promise<OwnerStatement> {
  const query = new URLSearchParams({ from, to });
  return (await request<{ data: OwnerStatement }>(`/api/v1/accounting/owners/${ownerId}/statement?${query.toString()}`)).data;
}
