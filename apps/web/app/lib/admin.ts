export type OrganizationSettings = {
  organizationId: string;
  name: string;
  slug: string;
  status: string;
  timezone: string;
  defaultCurrency: string;
  countryCode: string;
  billingEmail?: string;
  updatedAt: string;
};

export type OrganizationMember = {
  userId: string;
  email: string;
  displayName: string;
  userStatus: string;
  role: "admin" | "manager" | "accountant" | "maintenance" | "viewer" | "owner" | "tenant" | "agent";
  createdAt: string;
};

export type PortalLink = {
  kind: "owner" | "tenant";
  userId: string;
  email: string;
  displayName: string;
  resourceId: string;
  resourceName: string;
};

export type CurrencyAmount = { currency: string; amountMinor: number };

export type ReportingDashboard = {
  propertyCount: number;
  unitCount: number;
  occupiedUnits: number;
  vacancyRateBps: number;
  openMaintenance: number;
  emergencyMaintenance: number;
  leasesExpiring30Days: number;
  leasesExpiring90Days: number;
  outstandingByCurrency: CurrencyAmount[];
  overdueByCurrency: CurrencyAmount[];
  collectedThisMonthByCurrency: CurrencyAmount[];
};

export type AuditEvent = {
  id: string;
  actorUserId?: string;
  actorName?: string;
  action: string;
  resourceType: string;
  resourceId?: string;
  requestId?: string;
  metadata: Record<string, unknown>;
  occurredAt: string;
};

const apiBaseUrl = process.env.API_BASE_URL ?? "http://localhost:8080";
const organizationId = process.env.PROPERTY_OS_ORGANIZATION_ID;
const userId = process.env.PROPERTY_OS_USER_ID;
const accessToken = process.env.PROPERTY_OS_ACCESS_TOKEN;

export class AdminConfigurationError extends Error {}

function headers(): HeadersInit {
  if (!organizationId || (!accessToken && !userId)) {
    throw new AdminConfigurationError(
      "Set PROPERTY_OS_ORGANIZATION_ID and either PROPERTY_OS_ACCESS_TOKEN or PROPERTY_OS_USER_ID.",
    );
  }
  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    "X-Organization-ID": organizationId,
    ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : { "X-User-ID": userId as string }),
  };
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers: { ...headers(), ...(init?.headers ?? {}) },
    cache: "no-store",
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? `Property OS API returned ${response.status}`);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export async function getOrganizationSettings(): Promise<OrganizationSettings> {
  return (await request<{ data: OrganizationSettings }>("/api/v1/organization")).data;
}
export async function updateOrganizationSettings(input: { name: string; timezone: string; defaultCurrency: string; countryCode: string; billingEmail?: string }): Promise<OrganizationSettings> {
  return (await request<{ data: OrganizationSettings }>("/api/v1/organization", { method: "PATCH", body: JSON.stringify(input) })).data;
}
export async function listOrganizationMembers(): Promise<OrganizationMember[]> {
  return (await request<{ data: OrganizationMember[] }>("/api/v1/organization/members")).data;
}
export async function createOrganizationMember(input: { email: string; displayName: string; role: OrganizationMember["role"] }): Promise<OrganizationMember> {
  return (await request<{ data: OrganizationMember }>("/api/v1/organization/members", { method: "POST", body: JSON.stringify(input) })).data;
}
export async function updateOrganizationMemberRole(userId: string, role: OrganizationMember["role"]): Promise<OrganizationMember> {
  return (await request<{ data: OrganizationMember }>(`/api/v1/organization/members/${userId}`, { method: "PATCH", body: JSON.stringify({ role }) })).data;
}
export async function removeOrganizationMember(userId: string): Promise<void> {
  await request<void>(`/api/v1/organization/members/${userId}`, { method: "DELETE" });
}
export async function listPortalLinks(): Promise<PortalLink[]> {
  return (await request<{ data: PortalLink[] }>("/api/v1/organization/portal-links")).data;
}
export async function linkOwnerPortal(userId: string, ownerId: string): Promise<PortalLink> {
  return (await request<{ data: PortalLink }>("/api/v1/organization/portal-links/owners", { method: "POST", body: JSON.stringify({ userId, ownerId }) })).data;
}
export async function linkTenantPortal(userId: string, tenantId: string): Promise<PortalLink> {
  return (await request<{ data: PortalLink }>("/api/v1/organization/portal-links/tenants", { method: "POST", body: JSON.stringify({ userId, tenantId }) })).data;
}
export async function unlinkOwnerPortal(userId: string, ownerId: string): Promise<void> {
  await request<void>(`/api/v1/organization/portal-links/owners/${ownerId}/users/${userId}`, { method: "DELETE" });
}
export async function unlinkTenantPortal(userId: string, tenantId: string): Promise<void> {
  await request<void>(`/api/v1/organization/portal-links/tenants/${tenantId}/users/${userId}`, { method: "DELETE" });
}
export async function getReportingDashboard(): Promise<ReportingDashboard> {
  return (await request<{ data: ReportingDashboard }>("/api/v1/reporting/dashboard")).data;
}
export async function listAuditEvents(limit = 100): Promise<AuditEvent[]> {
  return (await request<{ data: AuditEvent[] }>(`/api/v1/audit-events?limit=${limit}`)).data;
}
