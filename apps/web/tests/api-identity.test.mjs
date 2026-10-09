import assert from "node:assert/strict";
import test from "node:test";
import { resolveAPIIdentity, APIIdentityError } from "../app/lib/api-identity.ts";

const shared = {
  PROPERTY_OS_ORGANIZATION_ID: "org-a",
  PROPERTY_OS_USER_ID: "admin-a",
  PROPERTY_OS_ACCESS_TOKEN: "shared-admin-token",
  PROPERTY_OS_OWNER_PORTAL_USER_ID: "owner-a",
  PROPERTY_OS_TENANT_PORTAL_USER_ID: "tenant-a",
};

test("production rejects missing caller token even when shared credentials exist", () => {
  for (const APP_ENV of ["production", "staging", "unknown"]) {
    assert.throws(() => resolveAPIIdentity({ ...shared, APP_ENV }, {}), APIIdentityError);
  }
  assert.throws(() => resolveAPIIdentity({ ...shared, NODE_ENV: "production" }, {}), APIIdentityError);
  assert.throws(() => resolveAPIIdentity(shared, {}), APIIdentityError);
});

test("production forwards caller token and requested organization without a development user", () => {
  assert.deepEqual(resolveAPIIdentity({ ...shared, APP_ENV: "production" }, {
    authorization: "Bearer user-specific-token", organizationId: "org-b",
  }, "owner"), { Authorization: "Bearer user-specific-token", "X-Organization-ID": "org-b" });
});

test("malformed caller authorization cannot fall back to an environment admin", () => {
  for (const authorization of ["", "Basic credentials", "Bearer", "Bearer two tokens", "Bearer token\n"]) {
    assert.throws(() => resolveAPIIdentity({ ...shared, APP_ENV: "development" }, { authorization }), APIIdentityError);
  }
});

test("development token and portal identities remain available only in explicit development/test", () => {
  assert.equal(resolveAPIIdentity({ ...shared, APP_ENV: "test" }, {}).Authorization, "Bearer shared-admin-token");
  const env = { ...shared, APP_ENV: "development", PROPERTY_OS_ACCESS_TOKEN: undefined };
  assert.equal(resolveAPIIdentity(env, {})["X-User-ID"], "admin-a");
  assert.equal(resolveAPIIdentity(env, {}, "owner")["X-User-ID"], "owner-a");
  assert.equal(resolveAPIIdentity(env, {}, "tenant")["X-User-ID"], "tenant-a");
});

test("caller token takes priority over configured development token", () => {
  assert.equal(resolveAPIIdentity({ ...shared, APP_ENV: "development" }, { authorization: "Bearer caller-token" }).Authorization, "Bearer caller-token");
});

test("missing organization and unconfigured development identity fail closed", () => {
  assert.throws(() => resolveAPIIdentity({ APP_ENV: "production" }, { authorization: "Bearer caller-token" }), APIIdentityError);
  assert.throws(() => resolveAPIIdentity({ APP_ENV: "test", PROPERTY_OS_ORGANIZATION_ID: "org-a" }, {}), APIIdentityError);
});
