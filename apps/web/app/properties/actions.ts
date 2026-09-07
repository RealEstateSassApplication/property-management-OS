"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { createProperty, createUnit } from "../lib/property-os";

function optionalNumber(value: FormDataEntryValue | null): number | undefined {
  if (typeof value !== "string" || value.trim() === "") return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

export async function createPropertyAction(formData: FormData) {
  const property = await createProperty({
    referenceCode: String(formData.get("referenceCode") ?? "").trim() || undefined,
    name: String(formData.get("name") ?? ""),
    propertyType: String(formData.get("propertyType") ?? "building"),
    addressLine1: String(formData.get("addressLine1") ?? ""),
    city: String(formData.get("city") ?? ""),
    region: String(formData.get("region") ?? "").trim() || undefined,
    countryCode: String(formData.get("countryCode") ?? "LK"),
  });

  revalidatePath("/properties");
  redirect(`/properties/${property.id}`);
}

export async function createUnitAction(formData: FormData) {
  const propertyId = String(formData.get("propertyId") ?? "");
  const floorAreaUnitRaw = String(formData.get("floorAreaUnit") ?? "");
  const floorAreaUnit = floorAreaUnitRaw === "sqft" || floorAreaUnitRaw === "sqm" ? floorAreaUnitRaw : undefined;

  await createUnit(propertyId, {
    referenceCode: String(formData.get("referenceCode") ?? ""),
    label: String(formData.get("label") ?? ""),
    bedrooms: optionalNumber(formData.get("bedrooms")),
    bathrooms: optionalNumber(formData.get("bathrooms")),
    floorArea: optionalNumber(formData.get("floorArea")),
    floorAreaUnit,
  });

  revalidatePath(`/properties/${propertyId}`);
}
