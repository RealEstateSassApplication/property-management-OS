"use server";

import { revalidatePath } from "next/cache";
import { createTenant } from "../lib/property-os";

export async function createTenantAction(formData: FormData) {
  const legalName = String(formData.get("legalName") ?? "").trim();
  const email = String(formData.get("email") ?? "").trim();
  const phone = String(formData.get("phone") ?? "").trim();
  const status = String(formData.get("status") ?? "prospect") as
    | "prospect"
    | "active"
    | "former"
    | "blocked";

  await createTenant({ legalName, email, phone, status });
  revalidatePath("/tenants");
  revalidatePath("/leases");
}
