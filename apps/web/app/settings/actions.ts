"use server";

import { revalidatePath } from "next/cache";
import {
  createOrganizationMember,
  linkOwnerPortal,
  linkTenantPortal,
  removeOrganizationMember,
  unlinkOwnerPortal,
  unlinkTenantPortal,
  updateOrganizationMemberRole,
  updateOrganizationSettings,
  type OrganizationMember,
} from "../lib/admin";

export async function updateOrganizationAction(formData: FormData) {
  await updateOrganizationSettings({
    name: String(formData.get("name") ?? "").trim(),
    timezone: String(formData.get("timezone") ?? "Asia/Colombo").trim(),
    defaultCurrency: String(formData.get("defaultCurrency") ?? "LKR").trim().toUpperCase(),
    countryCode: String(formData.get("countryCode") ?? "LK").trim().toUpperCase(),
    billingEmail: String(formData.get("billingEmail") ?? "").trim(),
  });
  revalidatePath("/settings");
}

export async function createMemberAction(formData: FormData) {
  await createOrganizationMember({
    email: String(formData.get("email") ?? "").trim(),
    displayName: String(formData.get("displayName") ?? "").trim(),
    role: String(formData.get("role") ?? "viewer") as OrganizationMember["role"],
  });
  revalidatePath("/settings");
}

export async function updateMemberRoleAction(formData: FormData) {
  await updateOrganizationMemberRole(
    String(formData.get("userId") ?? "").trim(),
    String(formData.get("role") ?? "viewer") as OrganizationMember["role"],
  );
  revalidatePath("/settings");
}

export async function removeMemberAction(formData: FormData) {
  await removeOrganizationMember(String(formData.get("userId") ?? "").trim());
  revalidatePath("/settings");
}

export async function linkOwnerPortalAction(formData: FormData) {
  await linkOwnerPortal(
    String(formData.get("userId") ?? "").trim(),
    String(formData.get("ownerId") ?? "").trim(),
  );
  revalidatePath("/settings");
}

export async function linkTenantPortalAction(formData: FormData) {
  await linkTenantPortal(
    String(formData.get("userId") ?? "").trim(),
    String(formData.get("tenantId") ?? "").trim(),
  );
  revalidatePath("/settings");
}

export async function unlinkOwnerPortalAction(formData: FormData) {
  await unlinkOwnerPortal(
    String(formData.get("userId") ?? "").trim(),
    String(formData.get("resourceId") ?? "").trim(),
  );
  revalidatePath("/settings");
}

export async function unlinkTenantPortalAction(formData: FormData) {
  await unlinkTenantPortal(
    String(formData.get("userId") ?? "").trim(),
    String(formData.get("resourceId") ?? "").trim(),
  );
  revalidatePath("/settings");
}
