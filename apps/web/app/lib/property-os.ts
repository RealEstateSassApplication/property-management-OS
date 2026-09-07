export type Property = {
  id: string;
  organizationId: string;
  referenceCode?: string;
  name: string;
  propertyType: string;
  addressLine1: string;
  addressLine2?: string;
  city: string;
  region?: string;
  postalCode?: string;
  countryCode: string;
  status: "active" | "inactive" | "archived";
  externalAvaraPropertyId?: string;
  createdAt: string;
  updatedAt: string;
};

export type Unit = {
  id: string;
  organizationId: string;
  propertyId: string;
  referenceCode: string;
  label: string;
  bedrooms?: number;
  bathrooms?: number;
  floorArea?: number;
  floorAreaUnit?: "sqft" | "sqm";
  occupancyStatus: "vacant" | "occupied" | "reserved" | "unavailable";
  createdAt: string;
  updatedAt: string;
};

const apiBaseUrl = process.env.API_BASE_URL ?? process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";
const organizationId = process.env.PROPERTY_OS_ORGANIZATION_ID;

export class PropertyOSConfigurationError extends Error {}

function headers(extra?: HeadersInit): HeadersInit {
  if (!organizationId) {
    throw new PropertyOSConfigurationError(
      "Set PROPERTY_OS_ORGANIZATION_ID for local development.",
    );
  }

  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    "X-Organization-ID": organizationId,
    ...extra,
  };
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers: headers(init?.headers),
    cache: "no-store",
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => null) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? `Property OS API returned ${response.status}`);
  }

  return response.json() as Promise<T>;
}

export async function listProperties(): Promise<Property[]> {
  const response = await request<{ data: Property[] }>("/api/v1/properties");
  return response.data;
}

export async function getProperty(id: string): Promise<Property> {
  const response = await request<{ data: Property }>(`/api/v1/properties/${id}`);
  return response.data;
}

export async function createProperty(input: {
  referenceCode?: string;
  name: string;
  propertyType: string;
  addressLine1: string;
  city: string;
  region?: string;
  countryCode: string;
}): Promise<Property> {
  const response = await request<{ data: Property }>("/api/v1/properties", {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.data;
}

export async function listUnits(propertyId: string): Promise<Unit[]> {
  const response = await request<{ data: Unit[] }>(`/api/v1/properties/${propertyId}/units`);
  return response.data;
}

export async function createUnit(propertyId: string, input: {
  referenceCode: string;
  label: string;
  bedrooms?: number;
  bathrooms?: number;
  floorArea?: number;
  floorAreaUnit?: "sqft" | "sqm";
}): Promise<Unit> {
  const response = await request<{ data: Unit }>(`/api/v1/properties/${propertyId}/units`, {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.data;
}
