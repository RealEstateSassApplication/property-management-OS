"use server";

import { revalidatePath } from "next/cache";
import { createOwner, createOwnershipInterest } from "../lib/property-os";

function percentageToBps(raw: string): number {
  const normalized = raw.trim();
  if (!/^\d{1,3}(\.\d{1,2})?$/.test(normalized)) {
    throw new Error("Ownership must be a percentage with at most two decimal places.");
  }
  const [whole, decimals = ""] = normalized.split(".");
  const bps = Number.parseInt(whole, 10) * 100 + Number.parseInt((decimals + "00").slice(0, 2), 10);
  if (bps < 1 || bps > 10000) {
    throw new Error("Ownership must be greater than 0% and at most 100%.");
  }
  return bps;
}

export async function createOwnerAction(formData: FormData) {
  await createOwner({
    legalName: String(formData.get("legalName") ?? "").trim(),
    ownerType: String(formData.get("ownerType") ?? "individual") as "individual" | "company",
    email: String(formData.get("email") ?? "").trim(),
    phone: String(formData.get("phone") ?? "").trim(),
    status: "active",
  });
  revalidatePath("/owners");
}

export async function createOwnershipInterestAction(formData: FormData) {
  await createOwnershipInterest({
    ownerId: String(formData.get("ownerId") ?? "").trim(),
    propertyId: String(formData.get("propertyId") ?? "").trim(),
    ownershipBps: percentageToBps(String(formData.get("ownershipPercent") ?? "")),
    effectiveFrom: String(formData.get("effectiveFrom") ?? "").trim(),
  });
  revalidatePath("/owners");
  revalidatePath("/properties");
}
