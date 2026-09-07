import { AppShell } from "../components/app-shell";
import { listTenants, PropertyOSConfigurationError, type Tenant } from "../lib/property-os";
import { createTenantAction } from "./actions";

export const dynamic = "force-dynamic";

export default async function TenantsPage() {
  let tenants: Tenant[] = [];
  let errorMessage = "";

  try {
    tenants = await listTenants();
  } catch (error) {
    errorMessage =
      error instanceof PropertyOSConfigurationError
        ? error.message
        : "The tenant register could not be loaded. Confirm the Go API and PostgreSQL are running.";
  }

  const active = tenants.filter((tenant) => tenant.status === "active").length;
  const prospects = tenants.filter((tenant) => tenant.status === "prospect").length;

  return (
    <AppShell section="Tenants">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">PEOPLE / TENANT REGISTER</p>
          <h1 className="pageTitle">Tenants</h1>
          <p className="pageIntro">
            Keep renter records separate from portal identities, then attach them to occupancies and leases as the relationship progresses.
          </p>
        </div>
        <div className="metricStrip" aria-label="Tenant metrics">
          <div><strong>{tenants.length}</strong><span>Total</span></div>
          <div><strong>{active}</strong><span>Active</span></div>
          <div><strong>{prospects}</strong><span>Prospects</span></div>
        </div>
      </header>

      {errorMessage ? (
        <div className="noticePanel">
          <strong>Tenant data is unavailable</strong>
          <p>{errorMessage}</p>
        </div>
      ) : null}

      <section className="contentGrid">
        <article className="panel panelWide">
          <header className="panelHeader">
            <div>
              <p className="panelKicker">REGISTER</p>
              <h2>People</h2>
            </div>
            <span className="countBadge">{tenants.length}</span>
          </header>

          {tenants.length === 0 ? (
            <div className="emptyState">
              <h3>No tenants yet</h3>
              <p>Create the first renter or prospect. A tenancy can be created once a unit is selected.</p>
            </div>
          ) : (
            <div className="dataTableWrap">
              <table className="dataTable">
                <thead>
                  <tr><th>Tenant</th><th>Contact</th><th>Status</th></tr>
                </thead>
                <tbody>
                  {tenants.map((tenant) => (
                    <tr key={tenant.id}>
                      <td>
                        <strong>{tenant.legalName}</strong>
                        <span className="recordMeta">{tenant.id.slice(0, 8)}</span>
                      </td>
                      <td>
                        {tenant.email || tenant.phone || "—"}
                        {tenant.email && tenant.phone ? <span className="recordMeta">{tenant.phone}</span> : null}
                      </td>
                      <td><span className={`statusPill status-${tenant.status}`}>{tenant.status}</span></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </article>

        <aside className="panel">
          <header className="panelHeader">
            <div>
              <p className="panelKicker">NEW RECORD</p>
              <h2>Add tenant</h2>
            </div>
          </header>
          <form action={createTenantAction} className="stackedForm">
            <label>Legal name<input name="legalName" required placeholder="Maya Silva" /></label>
            <label>Email<input name="email" type="email" placeholder="maya@example.com" /></label>
            <label>Phone<input name="phone" placeholder="+94 77 555 0101" /></label>
            <label>
              Status
              <select name="status" defaultValue="prospect">
                <option value="prospect">Prospect</option>
                <option value="active">Active</option>
                <option value="former">Former</option>
                <option value="blocked">Blocked</option>
              </select>
            </label>
            <button className="primaryButton" type="submit">Create tenant</button>
          </form>
        </aside>
      </section>
    </AppShell>
  );
}
