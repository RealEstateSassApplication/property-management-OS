import { apiRequestHeaders } from "./request-identity";

export type Inspection = {
  id: string;
  organizationId: string;
  tenancyId: string;
  propertyId: string;
  propertyName: string;
  unitId: string;
  unitLabel: string;
  primaryTenantName: string;
  inspectionType: "move_in" | "move_out" | "periodic";
  status: "draft" | "in_progress" | "completed" | "acknowledged" | "cancelled";
  scheduledFor?: string;
  summary?: string;
  completedAt?: string;
  acknowledgedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type InspectionItem = {
  id: string;
  inspectionId: string;
  area: string;
  itemName: string;
  condition: "good" | "fair" | "poor" | "damaged" | "not_applicable";
  notes?: string;
  evidenceDocumentId?: string;
};

const apiBaseUrl = process.env.API_BASE_URL ?? "http://localhost:8080";

export class InspectionConfigurationError extends Error {}

async function headers(): Promise<Record<string, string>> {
  try {
    const identity = await apiRequestHeaders();
    return identity;
  } catch (error) {
    throw new InspectionConfigurationError(error instanceof Error ? error.message : "Property OS authentication is not configured.");
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, { ...init, headers: { ...await headers(), ...(init?.headers ?? {}) }, cache: "no-store" });
  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? `Property OS API returned ${response.status}`);
  }
  return response.json() as Promise<T>;
}

export async function listInspections(): Promise<Inspection[]> {
  return (await request<{ data: Inspection[] }>("/api/v1/inspections")).data;
}

export async function listInspectionItems(inspectionId: string): Promise<InspectionItem[]> {
  return (await request<{ data: InspectionItem[] }>(`/api/v1/inspections/${inspectionId}/items`)).data;
}

export async function createInspection(input: { tenancyId: string; inspectionType: string; scheduledFor?: string; summary?: string }): Promise<Inspection> {
  return (await request<{ data: Inspection }>("/api/v1/inspections", { method: "POST", body: JSON.stringify(input) })).data;
}

export async function createInspectionItem(inspectionId: string, input: { area: string; itemName: string; condition: string; notes?: string; evidenceDocumentId?: string }): Promise<InspectionItem> {
  return (await request<{ data: InspectionItem }>(`/api/v1/inspections/${inspectionId}/items`, { method: "POST", body: JSON.stringify(input) })).data;
}

export async function completeInspection(inspectionId: string, summary?: string): Promise<Inspection> {
  return (await request<{ data: Inspection }>(`/api/v1/inspections/${inspectionId}/complete`, { method: "POST", body: JSON.stringify({ summary: summary ?? "" }) })).data;
}

export async function acknowledgeInspection(inspectionId: string): Promise<Inspection> {
  return (await request<{ data: Inspection }>(`/api/v1/inspections/${inspectionId}/acknowledge`, { method: "POST", body: "{}" })).data;
}
