"use server";

import { revalidatePath } from "next/cache";
import { createLease, createTenancy } from "../lib/property-os";

function toMinorUnits(raw: string): number {
  const normalized = raw.replaceAll(",", "").trim();
  if (!/^\d+(\.\d{1,2})?$/.test(normalized)) {
    throw new Error("Amount must be a positive number with at most two decimal places.");
  }
  const [whole, decimals = ""] = normalized.split(".");
  const value = BigInt(whole) * 100n + BigInt((decimals + "00").slice(0, 2));
  if (value > BigInt(Number.MAX_SAFE_INTEGER)) {
    throw new Error("Amount is too large.");
  }
  return Number(value);
}

export async function createTenancyAction(formData: FormData) {
  const unitId = String(formData.get("unitId") ?? "").trim();
  const primaryTenantId = String(formData.get("primaryTenantId") ?? "").trim();
  const startDate = String(formData.get("startDate") ?? "").trim();
  const endDate = String(formData.get("endDate") ?? "").trim();
  const status = String(formData.get("status") ?? "upcoming") as "upcoming" | "active";

  await createTenancy({ unitId, primaryTenantId, startDate, endDate, status });
  revalidatePath("/leases");
  revalidatePath("/properties");
  revalidatePath("/tenants");
}

export async function createLeaseAction(formData: FormData) {
  const tenancyId = String(formData.get("tenancyId") ?? "").trim();
  const referenceCode = String(formData.get("referenceCode") ?? "").trim();
  const startDate = String(formData.get("startDate") ?? "").trim();
  const endDate = String(formData.get("endDate") ?? "").trim();
  const rentAmountMinor = toMinorUnits(String(formData.get("rentAmount") ?? ""));
  const depositRaw = String(formData.get("depositAmount") ?? "0");
  const depositAmountMinor = toMinorUnits(depositRaw || "0");
  const currency = String(formData.get("currency") ?? "LKR").trim().toUpperCase();
  const dueDay = Number.parseInt(String(formData.get("dueDay") ?? "1"), 10);
  const status = String(formData.get("status") ?? "draft") as "draft" | "active";

  await createLease({
    tenancyId,
    referenceCode,
    startDate,
    endDate,
    rentAmountMinor,
    depositAmountMinor,
    currency,
    dueDay,
    status,
  });
  revalidatePath("/leases");
  revalidatePath("/properties");
  revalidatePath("/tenants");
}
