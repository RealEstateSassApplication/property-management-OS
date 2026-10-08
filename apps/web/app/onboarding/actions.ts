"use server";

export type PreviewIssue = {
  row: number;
  field: string;
  code: string;
  message: string;
};

export type PreviewProperty = {
  row: number;
  referenceCode?: string;
  name: string;
  propertyType: string;
  addressLine1: string;
  city: string;
  region?: string;
  countryCode: string;
  externalAvaraPropertyId?: string;
};

export type PortfolioPreview = {
  totalRows: number;
  readyRows: number;
  invalidRows: number;
  previewRows: PreviewProperty[];
  issues: PreviewIssue[];
  canImport: boolean;
};

export type PreflightState = {
  status: "idle" | "ready" | "error";
  message: string;
  report?: PortfolioPreview;
};

const maxCSVBytes = 512 * 1024;

export async function previewPropertyCSV(_previous: PreflightState, formData: FormData): Promise<PreflightState> {
  const file = formData.get("file");
  if (!(file instanceof File) || file.size === 0) {
    return { status: "error", message: "Choose a nonempty CSV file." };
  }
  if (file.size > maxCSVBytes) {
    return { status: "error", message: "The CSV must be at most 512 KiB." };
  }
  if (!file.name.toLowerCase().endsWith(".csv")) {
    return { status: "error", message: "Only .csv files are supported. Export your spreadsheet as CSV first." };
  }
  const organizationId = process.env.PROPERTY_OS_ORGANIZATION_ID;
  const accessToken = process.env.PROPERTY_OS_ACCESS_TOKEN;
  const developmentUserId = process.env.PROPERTY_OS_USER_ID;
  if (!organizationId || (!accessToken && !developmentUserId)) {
    return { status: "error", message: "Property OS authentication is not configured." };
  }

  try {
    const response = await fetch(`${process.env.API_BASE_URL ?? "http://localhost:8080"}/api/v1/onboarding/properties/preview`, {
      method: "POST",
      headers: {
        "Content-Type": "text/csv",
        "X-Organization-ID": organizationId,
        ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : { "X-User-ID": developmentUserId as string }),
      },
      body: await file.arrayBuffer(),
      cache: "no-store",
    });
    if (!response.ok) {
      const result = (await response.json().catch(() => null)) as { error?: { message?: string } } | null;
      return { status: "error", message: result?.error?.message ?? `Preflight failed (HTTP ${response.status}).` };
    }
    const result = (await response.json()) as { data: PortfolioPreview };
    return {
      status: "ready",
      message: result.data.canImport
        ? "All supplied rows passed the initial validation. Nothing has been imported."
        : "Some rows need corrections. Nothing has been imported.",
      report: result.data,
    };
  } catch {
    return { status: "error", message: "Cannot reach Property OS. Check the backend and try again." };
  }
}
