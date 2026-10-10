import { apiRequestHeaders } from "./request-identity";

export type AgentActionRecord = {
  id: string;
  organizationId: string;
  proposedByUserId: string;
  reviewedByUserId?: string;
  actionType: "maintenance.quote.approve";
  resourceType: "maintenance_quote";
  resourceId: string;
  riskLevel: "low" | "medium" | "high";
  title: string;
  reasoning: string;
  payload: {
    quoteId: string;
    workOrderId: string;
    workOrderSummary: string;
    vendorId: string;
    vendorName: string;
    amountMinor: number;
    currency: string;
    scopeSummary: string;
    quoteStatus: string;
  };
  status: "proposed" | "approved" | "rejected" | "executed" | "failed";
  decisionReason?: string;
  result: Record<string, unknown>;
  lastError?: string;
  reviewedAt?: string;
  executedAt?: string;
  createdAt: string;
  updatedAt: string;
};

const apiBaseUrl = process.env.API_BASE_URL ?? "http://localhost:8080";

export class AgentActionConfigurationError extends Error {}

async function headers(): Promise<Record<string, string>> {
  try {
    const identity = await apiRequestHeaders();
    return identity;
  } catch (error) {
    throw new AgentActionConfigurationError(error instanceof Error ? error.message : "Property OS authentication is not configured.");
  }
}

async function decode<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? `Property OS API returned ${response.status}`);
  }
  return response.json() as Promise<T>;
}

export async function listAgentActions(): Promise<AgentActionRecord[]> {
  const response = await fetch(`${apiBaseUrl}/api/v1/agent-actions`, {
    headers: await headers(),
    cache: "no-store",
  });
  return (await decode<{ data: AgentActionRecord[] }>(response)).data;
}

export async function decideAgentAction(id: string, decision: "approve" | "reject", reason: string): Promise<AgentActionRecord> {
  const response = await fetch(`${apiBaseUrl}/api/v1/agent-actions/${encodeURIComponent(id)}/decision`, {
    method: "POST",
    headers: await headers(),
    body: JSON.stringify({ decision, reason }),
    cache: "no-store",
  });
  return (await decode<{ data: AgentActionRecord }>(response)).data;
}
