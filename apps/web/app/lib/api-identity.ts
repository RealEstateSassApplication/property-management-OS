export class APIIdentityError extends Error {}

type IdentityEnvironment = Record<string, string | undefined>;
type IncomingIdentity = { authorization?: string | null; organizationId?: string | null };

// The API verifies the JWT and organization membership. These headers select
// a caller and scope; they never grant access on their own.
export function resolveAPIIdentity(
  env: IdentityEnvironment,
  incoming: IncomingIdentity,
  portal?: "owner" | "tenant",
): Record<string, string> {
  const organizationId = incoming.organizationId?.trim() || env.PROPERTY_OS_ORGANIZATION_ID?.trim();
  if (!organizationId) throw new APIIdentityError("Choose an organization before accessing Property OS.");
  const base = { "X-Organization-ID": organizationId };
  if (incoming.authorization != null) {
    if (!/^Bearer [^\s]+$/i.test(incoming.authorization)) {
      throw new APIIdentityError("A valid Bearer authorization header is required.");
    }
    return { ...base, Authorization: incoming.authorization };
  }
  const environment = env.APP_ENV ?? env.NODE_ENV;
  const development = environment === "development" || environment === "test";
  if (!development) {
    throw new APIIdentityError("Sign in through the configured OIDC gateway to access Property OS.");
  }
  if (env.PROPERTY_OS_ACCESS_TOKEN) {
    return { ...base, Authorization: `Bearer ${env.PROPERTY_OS_ACCESS_TOKEN}` };
  }
  const userId = (portal === "owner" ? env.PROPERTY_OS_OWNER_PORTAL_USER_ID
    : portal === "tenant" ? env.PROPERTY_OS_TENANT_PORTAL_USER_ID : undefined) || env.PROPERTY_OS_USER_ID;
  if (!userId) throw new APIIdentityError("Configure a development Property OS user or access token.");
  return { ...base, "X-User-ID": userId };
}
