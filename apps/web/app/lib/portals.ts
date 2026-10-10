import { apiRequestHeaders } from "./request-identity";

export type OwnerPortalSummary = {
  owners: Array<{ id: string; legalName: string; ownerType: string; email?: string; phone?: string }>;
  properties: Array<{
    ownerId: string;
    propertyId: string;
    referenceCode: string;
    name: string;
    city?: string;
    ownershipBps: number;
    unitCount: number;
    occupiedUnits: number;
    openMaintenance: number;
    outstandingByCurrencyMinor: Record<string, number>;
  }>;
};

export type TenantPortalSummary = {
  tenants: Array<{ id: string; legalName: string; email?: string; phone?: string; status: string }>;
  occupancies: Array<{
    tenantId: string;
    tenancyId: string;
    occupantRole: string;
    tenancyStatus: string;
    startDate: string;
    endDate?: string;
    propertyId: string;
    propertyName: string;
    unitId: string;
    unitLabel: string;
    leaseId?: string;
    leaseReference?: string;
    leaseStatus?: string;
    leaseStartDate?: string;
    leaseEndDate?: string;
    rentAmountMinor?: number;
    depositAmountMinor?: number;
    currency?: string;
    dueDay?: number;
  }>;
  rent: Array<{
    obligationId: string;
    tenantId: string;
    tenancyId: string;
    leaseId: string;
    period: string;
    dueDate: string;
    amountMinor: number;
    allocatedMinor: number;
    balanceMinor: number;
    currency: string;
    state: string;
  }>;
  maintenance: Array<{
    id: string;
    tenantId: string;
    propertyName: string;
    unitLabel?: string;
    title: string;
    category: string;
    priority: string;
    status: string;
    createdAt: string;
    updatedAt: string;
  }>;
};

const apiBaseUrl = process.env.API_BASE_URL ?? "http://localhost:8080";

export class PortalConfigurationError extends Error {}

async function headers(portal: "owner" | "tenant"): Promise<Record<string, string>> {
  try {
    const identity = await apiRequestHeaders(portal);
    return identity;
  } catch (error) {
    throw new PortalConfigurationError(error instanceof Error ? error.message : "Property OS authentication is not configured.");
  }
}

async function api<T>(path: string, portal: "owner" | "tenant", init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers: { ...await headers(portal), ...(init?.headers ?? {}) },
    cache: "no-store",
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? `Property OS API returned ${response.status}`);
  }
  return ((await response.json()) as { data: T }).data;
}

export function getOwnerPortalSummary(): Promise<OwnerPortalSummary> {
  return api<OwnerPortalSummary>("/api/v1/portal/owner", "owner");
}

export function getTenantPortalSummary(): Promise<TenantPortalSummary> {
  return api<TenantPortalSummary>("/api/v1/portal/tenant", "tenant");
}

export function createTenantPortalMaintenance(input: {
  tenancyId: string;
  title: string;
  description: string;
  category: string;
  priority: string;
}) {
  return api<{ id: string; status: string; title: string }>("/api/v1/portal/tenant/maintenance-requests", "tenant", {
    method: "POST",
    body: JSON.stringify(input),
  });
}
