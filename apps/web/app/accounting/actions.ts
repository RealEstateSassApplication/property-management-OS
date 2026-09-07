"use server";

import { revalidatePath } from "next/cache";
import {
  createDepositAccount,
  createDepositTransaction,
  createExpense,
  createRentAdjustment,
  reverseExpense,
  reversePayment,
} from "../lib/accounting";

function moneyToMinor(value: FormDataEntryValue | null): number {
  const text = String(value ?? "").trim();
  if (!/^\d+(?:\.\d{1,2})?$/.test(text)) throw new Error("Enter a valid positive amount with at most two decimal places.");
  const [whole, fraction = ""] = text.split(".");
  return Number(BigInt(whole) * 100n + BigInt((fraction + "00").slice(0, 2)));
}

export async function createAdjustmentAction(formData: FormData) {
  await createRentAdjustment({
    obligationId: String(formData.get("obligationId") ?? ""),
    adjustmentType: String(formData.get("adjustmentType") ?? ""),
    amountMinor: moneyToMinor(formData.get("amount")),
    reason: String(formData.get("reason") ?? ""),
  });
  revalidatePath("/accounting");
  revalidatePath("/rent");
}

export async function reversePaymentAction(formData: FormData) {
  await reversePayment(String(formData.get("paymentId") ?? ""), String(formData.get("reason") ?? ""));
  revalidatePath("/accounting");
  revalidatePath("/rent");
}

export async function createDepositAccountAction(formData: FormData) {
  await createDepositAccount(String(formData.get("leaseId") ?? ""));
  revalidatePath("/accounting");
}

export async function createDepositTransactionAction(formData: FormData) {
  await createDepositTransaction({
    depositAccountId: String(formData.get("depositAccountId") ?? ""),
    transactionType: String(formData.get("transactionType") ?? ""),
    amountMinor: moneyToMinor(formData.get("amount")),
    occurredOn: String(formData.get("occurredOn") ?? ""),
    note: String(formData.get("note") ?? ""),
  });
  revalidatePath("/accounting");
}

export async function createExpenseAction(formData: FormData) {
  await createExpense({
    propertyId: String(formData.get("propertyId") ?? ""),
    vendorId: String(formData.get("vendorId") ?? "") || undefined,
    workOrderId: String(formData.get("workOrderId") ?? "") || undefined,
    category: String(formData.get("category") ?? ""),
    amountMinor: moneyToMinor(formData.get("amount")),
    currency: String(formData.get("currency") ?? "LKR"),
    incurredOn: String(formData.get("incurredOn") ?? ""),
    note: String(formData.get("note") ?? ""),
    referenceCode: String(formData.get("referenceCode") ?? "") || undefined,
  });
  revalidatePath("/accounting");
}

export async function reverseExpenseAction(formData: FormData) {
  await reverseExpense(String(formData.get("expenseId") ?? ""), String(formData.get("reason") ?? ""));
  revalidatePath("/accounting");
}
