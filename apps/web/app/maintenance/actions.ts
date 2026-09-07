"use server";

import { revalidatePath } from "next/cache";
import {
  createMaintenanceEvidence,
  createMaintenanceQuote,
  createMaintenanceRequest,
  createMaintenanceVendor,
  createMaintenanceWorkOrder,
  decideMaintenanceQuote,
  updateMaintenanceRequestStatus,
  updateMaintenanceWorkOrderStatus,
  type MaintenanceRequest,
  type MaintenanceVendor,
  type MaintenanceWorkOrder,
} from "../lib/property-os";

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

function optional(value: FormDataEntryValue | null): string | undefined {
  const normalized = String(value ?? "").trim();
  return normalized || undefined;
}

export async function createVendorAction(formData: FormData) {
  await createMaintenanceVendor({
    name: String(formData.get("name") ?? "").trim(),
    trade: String(formData.get("trade") ?? "general") as MaintenanceVendor["trade"],
    email: optional(formData.get("email")),
    phone: optional(formData.get("phone")),
  });
  revalidatePath("/maintenance");
}

export async function createRequestAction(formData: FormData) {
  await createMaintenanceRequest({
    propertyId: String(formData.get("propertyId") ?? "").trim(),
    unitId: optional(formData.get("unitId")),
    tenantId: optional(formData.get("tenantId")),
    title: String(formData.get("title") ?? "").trim(),
    description: String(formData.get("description") ?? "").trim(),
    category: String(formData.get("category") ?? "other") as MaintenanceRequest["category"],
    priority: String(formData.get("priority") ?? "normal") as MaintenanceRequest["priority"],
  });
  revalidatePath("/maintenance");
}

export async function createWorkOrderAction(formData: FormData) {
  const scheduledRaw = optional(formData.get("scheduledFor"));
  await createMaintenanceWorkOrder({
    maintenanceRequestId: String(formData.get("maintenanceRequestId") ?? "").trim(),
    vendorId: optional(formData.get("vendorId")),
    summary: String(formData.get("summary") ?? "").trim(),
    scheduledFor: scheduledRaw ? new Date(scheduledRaw).toISOString() : undefined,
  });
  revalidatePath("/maintenance");
}

export async function createQuoteAction(formData: FormData) {
  await createMaintenanceQuote({
    workOrderId: String(formData.get("workOrderId") ?? "").trim(),
    vendorId: String(formData.get("vendorId") ?? "").trim(),
    amountMinor: toMinorUnits(String(formData.get("amount") ?? "")),
    currency: String(formData.get("currency") ?? "LKR").trim().toUpperCase(),
    scopeSummary: String(formData.get("scopeSummary") ?? "").trim(),
  });
  revalidatePath("/maintenance");
}

export async function decideQuoteAction(formData: FormData) {
  await decideMaintenanceQuote(
    String(formData.get("quoteId") ?? "").trim(),
    String(formData.get("decision") ?? "reject") as "approve" | "reject",
  );
  revalidatePath("/maintenance");
}

export async function createEvidenceAction(formData: FormData) {
  await createMaintenanceEvidence({
    workOrderId: String(formData.get("workOrderId") ?? "").trim(),
    evidenceType: "note",
    note: String(formData.get("note") ?? "").trim(),
  });
  revalidatePath("/maintenance");
}

export async function updateRequestStatusAction(formData: FormData) {
  await updateMaintenanceRequestStatus(
    String(formData.get("requestId") ?? "").trim(),
    String(formData.get("status") ?? "triaged") as MaintenanceRequest["status"],
  );
  revalidatePath("/maintenance");
}

export async function updateWorkOrderStatusAction(formData: FormData) {
  await updateMaintenanceWorkOrderStatus(
    String(formData.get("workOrderId") ?? "").trim(),
    String(formData.get("status") ?? "in_progress") as MaintenanceWorkOrder["status"],
  );
  revalidatePath("/maintenance");
}
