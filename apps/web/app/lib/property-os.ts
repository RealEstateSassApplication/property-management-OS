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

export type Tenant = {
  id: string;
  organizationId: string;
  legalName: string;
  email?: string;
  phone?: string;
  status: "prospect" | "active" | "former" | "blocked";
  createdAt: string;
  updatedAt: string;
};

export type Tenancy = {
  id: string;
  organizationId: string;
  unitId: string;
  unitLabel: string;
  propertyName: string;
  primaryTenantId: string;
  primaryTenantName: string;
  occupantCount: number;
  startDate: string;
  endDate?: string;
  status: "upcoming" | "active" | "ended" | "cancelled";
  createdAt: string;
  updatedAt: string;
};

export type Lease = {
  id: string;
  organizationId: string;
  tenancyId: string;
  referenceCode: string;
  propertyName: string;
  unitLabel: string;
  primaryTenantName: string;
  startDate: string;
  endDate: string;
  rentAmountMinor: number;
  depositAmountMinor: number;
  currency: string;
  dueDay: number;
  status: "draft" | "active" | "expired" | "terminated" | "cancelled";
  createdAt: string;
  updatedAt: string;
};

const apiBaseUrl = process.env.API_BASE_URL ?? "http://localhost:8080";
const organizationId = process.env.PROPERTY_OS_ORGANIZATION_ID;
const userId = process.env.PROPERTY_OS_USER_ID;

export class PropertyOSConfigurationError extends Error {}

function headers(extra?: HeadersInit): HeadersInit {
  if (!organizationId || !userId) {
    throw new PropertyOSConfigurationError(
      "Set PROPERTY_OS_ORGANIZATION_ID and PROPERTY_OS_USER_ID for local development.",
    );
  }

  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    "X-Organization-ID": organizationId,
    "X-User-ID": userId,
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
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
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

export async function createUnit(
  propertyId: string,
  input: {
    referenceCode: string;
    label: string;
    bedrooms?: number;
    bathrooms?: number;
    floorArea?: number;
    floorAreaUnit?: "sqft" | "sqm";
  },
): Promise<Unit> {
  const response = await request<{ data: Unit }>(`/api/v1/properties/${propertyId}/units`, {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.data;
}

export async function listTenants(): Promise<Tenant[]> {
  const response = await request<{ data: Tenant[] }>("/api/v1/tenants");
  return response.data;
}

export async function createTenant(input: {
  legalName: string;
  email?: string;
  phone?: string;
  status?: Tenant["status"];
}): Promise<Tenant> {
  const response = await request<{ data: Tenant }>("/api/v1/tenants", {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.data;
}

export async function listTenancies(): Promise<Tenancy[]> {
  const response = await request<{ data: Tenancy[] }>("/api/v1/tenancies");
  return response.data;
}

export async function createTenancy(input: {
  unitId: string;
  primaryTenantId: string;
  startDate: string;
  endDate?: string;
  status?: Tenancy["status"];
}): Promise<Tenancy> {
  const response = await request<{ data: Tenancy }>("/api/v1/tenancies", {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.data;
}

export async function listLeases(): Promise<Lease[]> {
  const response = await request<{ data: Lease[] }>("/api/v1/leases");
  return response.data;
}

export async function createLease(input: {
  tenancyId: string;
  referenceCode: string;
  startDate: string;
  endDate: string;
  rentAmountMinor: number;
  depositAmountMinor: number;
  currency: string;
  dueDay: number;
  status?: Lease["status"];
}): Promise<Lease> {
  const response = await request<{ data: Lease }>("/api/v1/leases", {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.data;
}
