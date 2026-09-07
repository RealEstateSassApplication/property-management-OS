"use server";

import { revalidatePath } from "next/cache";
import {
  acknowledgeInspection,
  completeInspection,
  createInspection,
  createInspectionItem,
} from "../lib/inspections";

export async function createInspectionAction(formData: FormData) {
  const tenancyId = String(formData.get("tenancyId") ?? "");
  const inspectionType = String(formData.get("inspectionType") ?? "periodic");
  const scheduledLocal = String(formData.get("scheduledFor") ?? "");
  const summary = String(formData.get("summary") ?? "");
  const scheduledFor = scheduledLocal ? new Date(scheduledLocal).toISOString() : undefined;
  await createInspection({ tenancyId, inspectionType, scheduledFor, summary });
  revalidatePath("/inspections");
}

export async function createInspectionItemAction(formData: FormData) {
  const inspectionId = String(formData.get("inspectionId") ?? "");
  await createInspectionItem(inspectionId, {
    area: String(formData.get("area") ?? ""),
    itemName: String(formData.get("itemName") ?? ""),
    condition: String(formData.get("condition") ?? "good"),
    notes: String(formData.get("notes") ?? ""),
    evidenceDocumentId: String(formData.get("evidenceDocumentId") ?? ""),
  });
  revalidatePath("/inspections");
}

export async function completeInspectionAction(formData: FormData) {
  await completeInspection(String(formData.get("inspectionId") ?? ""), String(formData.get("summary") ?? ""));
  revalidatePath("/inspections");
}

export async function acknowledgeInspectionAction(formData: FormData) {
  await acknowledgeInspection(String(formData.get("inspectionId") ?? ""));
  revalidatePath("/inspections");
}
