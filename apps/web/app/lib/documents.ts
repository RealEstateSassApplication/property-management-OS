export type StoredDocument = {
  id: string;
  organizationId: string;
  resourceType: string;
  resourceId: string;
  kind: string;
  fileName: string;
  contentType: string;
  sizeBytes: number;
  status: "pending" | "available" | "quarantined" | "deleted";
  checksumSha256?: string;
  uploadedByUserId?: string;
  verifiedAt?: string;
  deletedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type DocumentUploadIntent = {
  document: StoredDocument;
  uploadUrl: string;
  headers: Record<string, string>;
  expiresAt: string;
};

export type DocumentDownloadGrant = { url: string; expiresAt: string };

const apiBaseUrl = process.env.API_BASE_URL ?? "http://localhost:8080";
const organizationId = process.env.PROPERTY_OS_ORGANIZATION_ID;
const userId = process.env.PROPERTY_OS_USER_ID;

export class DocumentConfigurationError extends Error {}

function headers(): HeadersInit {
  if (!organizationId || !userId) {
    throw new DocumentConfigurationError(
      "Set PROPERTY_OS_ORGANIZATION_ID and PROPERTY_OS_USER_ID for the current development web session.",
    );
  }
  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    "X-Organization-ID": organizationId,
    "X-User-ID": userId,
  };
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers: { ...headers(), ...init?.headers },
    cache: "no-store",
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
    throw new Error(payload?.error?.message ?? `Property OS API returned ${response.status}`);
  }
  return response.json() as Promise<T>;
}

export async function listDocuments(): Promise<StoredDocument[]> {
  return (await request<{ data: StoredDocument[] }>("/api/v1/documents")).data;
}

export async function initiateDocumentUpload(input: {
  resourceType: string;
  resourceId: string;
  kind: string;
  fileName: string;
  contentType: string;
  sizeBytes: number;
}): Promise<DocumentUploadIntent> {
  return (
    await request<{ data: DocumentUploadIntent }>("/api/v1/documents/uploads", {
      method: "POST",
      body: JSON.stringify(input),
    })
  ).data;
}

export async function completeDocumentUpload(documentId: string): Promise<StoredDocument> {
  return (
    await request<{ data: StoredDocument }>(`/api/v1/documents/${documentId}/complete`, {
      method: "POST",
      body: "{}",
    })
  ).data;
}

export async function getDocumentDownload(documentId: string): Promise<DocumentDownloadGrant> {
  return (await request<{ data: DocumentDownloadGrant }>(`/api/v1/documents/${documentId}/download`)).data;
}

export async function deleteDocument(documentId: string): Promise<StoredDocument> {
  return (
    await request<{ data: StoredDocument }>(`/api/v1/documents/${documentId}`, {
      method: "DELETE",
    })
  ).data;
}
