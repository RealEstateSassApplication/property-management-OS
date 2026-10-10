import { apiRequestHeaders } from "./request-identity";

export type NotificationRecord = {
  id: string;
  organizationId: string;
  actorUserId?: string;
  topic: string;
  channel: "email" | "sms" | "whatsapp" | "webhook";
  recipient: string;
  subject?: string;
  body: string;
  resourceType?: string;
  resourceId?: string;
  status: "pending" | "processing" | "retry" | "delivered" | "dead";
  attemptCount: number;
  maxAttempts: number;
  availableAt: string;
  lastError?: string;
  deliveredAt?: string;
  createdAt: string;
  updatedAt: string;
};

const apiBaseUrl = process.env.API_BASE_URL ?? "http://localhost:8080";

export class NotificationConfigurationError extends Error {}

async function headers(): Promise<Record<string, string>> {
  try {
    const identity = await apiRequestHeaders();
    return identity;
  } catch (error) {
    throw new NotificationConfigurationError(error instanceof Error ? error.message : "Property OS authentication is not configured.");
  }
}

export async function listNotifications(): Promise<NotificationRecord[]> {
  const response = await fetch(`${apiBaseUrl}/api/v1/notifications`, {
    headers: await headers(),
    cache: "no-store",
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? `Property OS API returned ${response.status}`);
  }
  return ((await response.json()) as { data: NotificationRecord[] }).data;
}
