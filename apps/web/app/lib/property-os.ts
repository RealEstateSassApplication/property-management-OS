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

export type Owner = {
  id: string;
  organizationId: string;
  legalName: string;
  ownerType: "individual" | "company";
  email?: string;
  phone?: string;
  status: "active" | "inactive";
  createdAt: string;
  updatedAt: string;
};

export type OwnershipInterest = {
  id: string;
  organizationId: string;
  propertyId: string;
  propertyName: string;
  ownerId: string;
  ownerName: string;
  ownershipBps: number;
  effectiveFrom: string;
  effectiveTo?: string;
  createdAt: string;
};

export type RentObligation = {
  id: string;
  organizationId: string;
  leaseId: string;
  leaseReference: string;
  propertyName: string;
  unitLabel: string;
  primaryTenantName: string;
  period: string;
  dueDate: string;
  amountMinor: number;
  allocatedMinor: number;
  balanceMinor: number;
  currency: string;
  state: "open" | "overdue" | "paid" | "void";
  createdAt: string;
  updatedAt: string;
};

export type RentPayment = {
  id: string;
  organizationId: string;
  tenantId: string;
  tenantName: string;
  amountMinor: number;
  allocatedMinor: number;
  unallocatedMinor: number;
  currency: string;
  receivedAt: string;
  method: "cash" | "bank_transfer" | "card" | "online" | "other";
  referenceCode?: string;
  status: "posted" | "void";
  createdAt: string;
};

export type RentAllocation = {
  id: string;
  organizationId: string;
  paymentId: string;
  obligationId: string;
  amountMinor: number;
  createdAt: string;
};

export type MaintenanceVendor = {
  id: string;
  organizationId: string;
  name: string;
  trade: "plumbing" | "electrical" | "hvac" | "appliance" | "structural" | "cleaning" | "security" | "general" | "other";
  email?: string;
  phone?: string;
  status: "active" | "inactive";
  createdAt: string;
  updatedAt: string;
};

export type MaintenanceRequest = {
  id: string;
  organizationId: string;
  propertyId: string;
  propertyName: string;
  unitId?: string;
  unitLabel?: string;
  tenantId?: string;
  tenantName?: string;
  title: string;
  description: string;
  category: "plumbing" | "electrical" | "hvac" | "appliance" | "structural" | "cleaning" | "security" | "other";
  priority: "low" | "normal" | "high" | "emergency";
  status: "open" | "triaged" | "in_progress" | "resolved" | "cancelled";
  resolvedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type MaintenanceWorkOrder = {
  id: string;
  organizationId: string;
  maintenanceRequestId: string;
  requestTitle: string;
  propertyName: string;
  unitLabel?: string;
  vendorId?: string;
  vendorName?: string;
  assignedUserId?: string;
  summary: string;
  status: "planned" | "assigned" | "in_progress" | "completed" | "cancelled";
  scheduledFor?: string;
  startedAt?: string;
  completedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type MaintenanceQuote = {
  id: string;
  organizationId: string;
  workOrderId: string;
  workOrderSummary: string;
  vendorId: string;
  vendorName: string;
  amountMinor: number;
  currency: string;
  scopeSummary: string;
  status: "submitted" | "approved" | "rejected" | "withdrawn";
  submittedAt: string;
  reviewedAt?: string;
  reviewedByUserId?: string;
  createdAt: string;
  updatedAt: string;
};

export type MaintenanceEvidence = {
  id: string;
  organizationId: string;
  workOrderId: string;
  evidenceType: "note" | "photo" | "invoice" | "receipt" | "other";
  note?: string;
  storageKey?: string;
  submittedByUserId?: string;
  createdAt: string;
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
  return (await request<{ data: Property[] }>("/api/v1/properties")).data;
}

export async function getProperty(id: string): Promise<Property> {
  return (await request<{ data: Property }>(`/api/v1/properties/${id}`)).data;
}

export async function createProperty(input: { referenceCode?: string; name: string; propertyType: string; addressLine1: string; city: string; region?: string; countryCode: string }): Promise<Property> {
  return (await request<{ data: Property }>("/api/v1/properties", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function listUnits(propertyId: string): Promise<Unit[]> {
  return (await request<{ data: Unit[] }>(`/api/v1/properties/${propertyId}/units`)).data;
}

export async function createUnit(propertyId: string, input: { referenceCode: string; label: string; bedrooms?: number; bathrooms?: number; floorArea?: number; floorAreaUnit?: "sqft" | "sqm" }): Promise<Unit> {
  return (await request<{ data: Unit }>(`/api/v1/properties/${propertyId}/units`, { method: "POST", body: JSON.stringify(input) })).data;
}

export async function listTenants(): Promise<Tenant[]> {
  return (await request<{ data: Tenant[] }>("/api/v1/tenants")).data;
}

export async function createTenant(input: { legalName: string; email?: string; phone?: string; status?: Tenant["status"] }): Promise<Tenant> {
  return (await request<{ data: Tenant }>("/api/v1/tenants", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function listTenancies(): Promise<Tenancy[]> {
  return (await request<{ data: Tenancy[] }>("/api/v1/tenancies")).data;
}

export async function createTenancy(input: { unitId: string; primaryTenantId: string; startDate: string; endDate?: string; status?: Tenancy["status"] }): Promise<Tenancy> {
  return (await request<{ data: Tenancy }>("/api/v1/tenancies", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function listLeases(): Promise<Lease[]> {
  return (await request<{ data: Lease[] }>("/api/v1/leases")).data;
}

export async function createLease(input: { tenancyId: string; referenceCode: string; startDate: string; endDate: string; rentAmountMinor: number; depositAmountMinor: number; currency: string; dueDay: number; status?: Lease["status"] }): Promise<Lease> {
  return (await request<{ data: Lease }>("/api/v1/leases", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function listOwners(): Promise<Owner[]> {
  return (await request<{ data: Owner[] }>("/api/v1/owners")).data;
}

export async function createOwner(input: { legalName: string; ownerType: Owner["ownerType"]; email?: string; phone?: string; status?: Owner["status"] }): Promise<Owner> {
  return (await request<{ data: Owner }>("/api/v1/owners", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function listOwnershipInterests(): Promise<OwnershipInterest[]> {
  return (await request<{ data: OwnershipInterest[] }>("/api/v1/ownership-interests")).data;
}

export async function createOwnershipInterest(input: { ownerId: string; propertyId: string; ownershipBps: number; effectiveFrom: string }): Promise<OwnershipInterest> {
  return (await request<{ data: OwnershipInterest }>("/api/v1/ownership-interests", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function listRentObligations(): Promise<RentObligation[]> {
  return (await request<{ data: RentObligation[] }>("/api/v1/rent/obligations")).data;
}

export async function createRentObligation(input: { leaseId: string; period: string }): Promise<RentObligation> {
  return (await request<{ data: RentObligation }>("/api/v1/rent/obligations", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function listRentPayments(): Promise<RentPayment[]> {
  return (await request<{ data: RentPayment[] }>("/api/v1/rent/payments")).data;
}

export async function createRentPayment(input: { tenantId: string; amountMinor: number; currency: string; receivedAt: string; method: RentPayment["method"]; referenceCode?: string }): Promise<RentPayment> {
  return (await request<{ data: RentPayment }>("/api/v1/rent/payments", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function createRentAllocation(input: { paymentId: string; obligationId: string; amountMinor: number }): Promise<RentAllocation> {
  return (await request<{ data: RentAllocation }>("/api/v1/rent/allocations", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function listMaintenanceVendors(): Promise<MaintenanceVendor[]> {
  return (await request<{ data: MaintenanceVendor[] }>("/api/v1/maintenance/vendors")).data;
}

export async function createMaintenanceVendor(input: { name: string; trade: MaintenanceVendor["trade"]; email?: string; phone?: string }): Promise<MaintenanceVendor> {
  return (await request<{ data: MaintenanceVendor }>("/api/v1/maintenance/vendors", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function listMaintenanceRequests(): Promise<MaintenanceRequest[]> {
  return (await request<{ data: MaintenanceRequest[] }>("/api/v1/maintenance/requests")).data;
}

export async function createMaintenanceRequest(input: { propertyId: string; unitId?: string; tenantId?: string; title: string; description: string; category: MaintenanceRequest["category"]; priority: MaintenanceRequest["priority"] }): Promise<MaintenanceRequest> {
  return (await request<{ data: MaintenanceRequest }>("/api/v1/maintenance/requests", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function updateMaintenanceRequestStatus(id: string, status: MaintenanceRequest["status"]): Promise<MaintenanceRequest> {
  return (await request<{ data: MaintenanceRequest }>(`/api/v1/maintenance/requests/${id}/status`, { method: "PATCH", body: JSON.stringify({ status }) })).data;
}

export async function listMaintenanceWorkOrders(): Promise<MaintenanceWorkOrder[]> {
  return (await request<{ data: MaintenanceWorkOrder[] }>("/api/v1/maintenance/work-orders")).data;
}

export async function createMaintenanceWorkOrder(input: { maintenanceRequestId: string; vendorId?: string; summary: string; scheduledFor?: string }): Promise<MaintenanceWorkOrder> {
  return (await request<{ data: MaintenanceWorkOrder }>("/api/v1/maintenance/work-orders", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function updateMaintenanceWorkOrderStatus(id: string, status: MaintenanceWorkOrder["status"]): Promise<MaintenanceWorkOrder> {
  return (await request<{ data: MaintenanceWorkOrder }>(`/api/v1/maintenance/work-orders/${id}/status`, { method: "PATCH", body: JSON.stringify({ status }) })).data;
}

export async function listMaintenanceQuotes(): Promise<MaintenanceQuote[]> {
  return (await request<{ data: MaintenanceQuote[] }>("/api/v1/maintenance/quotes")).data;
}

export async function createMaintenanceQuote(input: { workOrderId: string; vendorId: string; amountMinor: number; currency: string; scopeSummary: string }): Promise<MaintenanceQuote> {
  return (await request<{ data: MaintenanceQuote }>("/api/v1/maintenance/quotes", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function decideMaintenanceQuote(id: string, decision: "approve" | "reject"): Promise<MaintenanceQuote> {
  return (await request<{ data: MaintenanceQuote }>(`/api/v1/maintenance/quotes/${id}/decision`, { method: "POST", body: JSON.stringify({ decision }) })).data;
}

export async function listMaintenanceEvidence(): Promise<MaintenanceEvidence[]> {
  return (await request<{ data: MaintenanceEvidence[] }>("/api/v1/maintenance/evidence")).data;
}

export async function createMaintenanceEvidence(input: { workOrderId: string; evidenceType: MaintenanceEvidence["evidenceType"]; note?: string; storageKey?: string }): Promise<MaintenanceEvidence> {
  return (await request<{ data: MaintenanceEvidence }>("/api/v1/maintenance/evidence", { method: "POST", body: JSON.stringify(input) })).data;
}
