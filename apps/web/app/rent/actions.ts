"use server";

import { revalidatePath } from "next/cache";
import { createRentAllocation, createRentObligation, createRentPayment } from "../lib/property-os";

function toMinorUnits(raw: string): number {
  const normalized = raw.replaceAll(",", "").trim();
  if (!/^\d+(\.\d{1,2})?$/.test(normalized)) {
    throw new Error("Amount must be a positive number with at most two decimal places.");
  }
  const [whole, decimals = ""] = normalized.split(".");
  const value = BigInt(whole) * 100n + BigInt((decimals + "00").slice(0, 2));
  if (value <= 0n || value > BigInt(Number.MAX_SAFE_INTEGER)) {
    throw new Error("Amount is outside the supported range.");
  }
  return Number(value);
}

export async function createRentObligationAction(formData: FormData) {
  await createRentObligation({
    leaseId: String(formData.get("leaseId") ?? "").trim(),
    period: String(formData.get("period") ?? "").trim(),
  });
  revalidatePath("/rent");
}

export async function createRentPaymentAction(formData: FormData) {
  await createRentPayment({
    tenantId: String(formData.get("tenantId") ?? "").trim(),
    amountMinor: toMinorUnits(String(formData.get("amount") ?? "")),
    currency: String(formData.get("currency") ?? "LKR").trim().toUpperCase(),
    receivedAt: String(formData.get("receivedAt") ?? "").trim(),
    method: String(formData.get("method") ?? "bank_transfer") as "cash" | "bank_transfer" | "card" | "online" | "other",
    referenceCode: String(formData.get("referenceCode") ?? "").trim(),
  });
  revalidatePath("/rent");
}

export async function createRentAllocationAction(formData: FormData) {
  await createRentAllocation({
    paymentId: String(formData.get("paymentId") ?? "").trim(),
    obligationId: String(formData.get("obligationId") ?? "").trim(),
    amountMinor: toMinorUnits(String(formData.get("amount") ?? "")),
  });
  revalidatePath("/rent");
}
