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
const organizationId = process.env.PROPERTY_OS_ORGANIZATION_ID;
const userId = process.env.PROPERTY_OS_USER_ID;
const accessToken = process.env.PROPERTY_OS_ACCESS_TOKEN;

export class NotificationConfigurationError extends Error {}

function headers(): HeadersInit {
  if (!organizationId || (!accessToken && !userId)) {
    throw new NotificationConfigurationError(
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

export async function listNotifications(): Promise<NotificationRecord[]> {
  const response = await fetch(`${apiBaseUrl}/api/v1/notifications`, {
    headers: headers(),
    cache: "no-store",
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? `Property OS API returned ${response.status}`);
  }
  return ((await response.json()) as { data: NotificationRecord[] }).data;
}
