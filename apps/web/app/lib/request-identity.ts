import { headers } from "next/headers";
import { resolveAPIIdentity } from "./api-identity";

export async function apiRequestHeaders(portal?: "owner" | "tenant"): Promise<Record<string, string>> {
  const incoming = await headers();
  return {
    Accept: "application/json",
    "Content-Type": "application/json",
    ...resolveAPIIdentity(process.env, {
      authorization: incoming.get("authorization"),
      organizationId: incoming.get("x-organization-id"),
    }, portal),
  };
}
