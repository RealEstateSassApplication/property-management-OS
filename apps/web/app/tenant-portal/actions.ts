"use server";

import { revalidatePath } from "next/cache";
import { createTenantPortalMaintenance } from "../lib/portals";

export async function createTenantMaintenanceAction(formData: FormData) {
  await createTenantPortalMaintenance({
    tenancyId: String(formData.get("tenancyId") ?? "").trim(),
    title: String(formData.get("title") ?? "").trim(),
    description: String(formData.get("description") ?? "").trim(),
    category: String(formData.get("category") ?? "other").trim(),
    priority: String(formData.get("priority") ?? "normal").trim(),
  });
  revalidatePath("/tenant-portal");
}
