import { PortalShell } from "../components/portal-shell";
import { getTenantPortalSummary, PortalConfigurationError, type TenantPortalSummary } from "../lib/portals";
import { createTenantMaintenanceAction } from "./actions";

export const dynamic = "force-dynamic";

function money(amountMinor: number, currency: string) {
  return new Intl.NumberFormat("en-LK", { style: "currency", currency, maximumFractionDigits: 2 }).format(amountMinor / 100);
}

export default async function TenantPortalPage() {
  let summary: TenantPortalSummary = { tenants: [], occupancies: [], rent: [], maintenance: [] };
  let errorMessage = "";
  try {
    summary = await getTenantPortalSummary();
  } catch (error) {
    errorMessage = error instanceof PortalConfigurationError ? error.message : "Tenant portal data could not be loaded.";
  }

  const activeOccupancies = summary.occupancies.filter((item) => item.tenancyStatus === "active");
  const outstanding = summary.rent.filter((item) => item.balanceMinor > 0 && item.state !== "void");
  const overdue = outstanding.filter((item) => item.state === "overdue");
  const activeMaintenance = summary.maintenance.filter((item) => !["resolved", "cancelled"].includes(item.status));
  const tenantName = summary.tenants[0]?.legalName ?? "Tenant";

  return (
    <PortalShell mode="tenant">
      <section className="portalHero">
        <div>
          <p className="eyebrow">TENANT VIEW</p>
          <h1 className="pageTitle">Welcome, {tenantName}.</h1>
          <p className="pageIntro">Your occupancy, lease, rent and maintenance history are loaded only through tenant records explicitly linked to your login.</p>
        </div>
      </section>

      {errorMessage ? <div className="noticePanel"><strong>Portal unavailable</strong><p>{errorMessage}</p></div> : null}

      <section className="metricStrip portalMetrics">
        <div><strong>{activeOccupancies.length}</strong><span>Active occupancy</span></div>
        <div><strong>{outstanding.length}</strong><span>Open rent items</span></div>
        <div><strong>{overdue.length}</strong><span>Overdue</span></div>
        <div><strong>{activeMaintenance.length}</strong><span>Open maintenance</span></div>
      </section>

      <div className="leasingGrid portalGrid">
        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">HOME & LEASE</p><h2>Current occupancy</h2></div></div>
          {activeOccupancies.length ? activeOccupancies.map((item) => (
            <div className="portalCard" key={item.tenancyId}>
              <strong>{item.propertyName} · {item.unitLabel}</strong>
              <span>{item.leaseReference || "Lease pending"}</span>
              <dl className="portalFacts">
                <div><dt>Lease status</dt><dd>{item.leaseStatus || "—"}</dd></div>
                <div><dt>Lease end</dt><dd>{item.leaseEndDate || "—"}</dd></div>
                <div><dt>Monthly rent</dt><dd>{item.currency && item.rentAmountMinor ? money(item.rentAmountMinor, item.currency) : "—"}</dd></div>
                <div><dt>Deposit</dt><dd>{item.currency && item.depositAmountMinor ? money(item.depositAmountMinor, item.currency) : "—"}</dd></div>
              </dl>
            </div>
          )) : <div className="emptyState"><h3>No active occupancy</h3><p>No active tenancy linked to this login is currently available.</p></div>}
        </section>

        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">SELF SERVICE</p><h2>Report maintenance</h2></div></div>
          <form className="stackedForm" action={createTenantMaintenanceAction}>
            <label>Current tenancy<select name="tenancyId" required defaultValue={activeOccupancies[0]?.tenancyId ?? ""}><option value="" disabled>Select occupancy</option>{activeOccupancies.map((item) => <option value={item.tenancyId} key={item.tenancyId}>{item.propertyName} · {item.unitLabel}</option>)}</select></label>
            <label>Issue title<input name="title" required placeholder="e.g. AC not cooling" /></label>
            <label>Description<input name="description" required placeholder="Describe what is happening" /></label>
            <div className="formPair">
              <label>Category<select name="category" defaultValue="other"><option>plumbing</option><option>electrical</option><option>hvac</option><option>appliance</option><option>structural</option><option>cleaning</option><option>security</option><option>other</option></select></label>
              <label>Priority<select name="priority" defaultValue="normal"><option>low</option><option>normal</option><option>high</option><option>emergency</option></select></label>
            </div>
            <button className="primaryButton" disabled={!activeOccupancies.length}>Create maintenance request</button>
          </form>
        </section>
      </div>

      <section className="panel portalPanel">
        <div className="panelHeader"><div><p className="panelKicker">RENT</p><h2>Rent history</h2></div><span className="countBadge">{summary.rent.length}</span></div>
        {summary.rent.length ? <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Period</th><th>Due</th><th>Amount</th><th>Paid</th><th>Balance</th><th>Status</th></tr></thead><tbody>{summary.rent.map((item) => <tr key={item.obligationId}><td>{item.period}</td><td>{item.dueDate}</td><td>{money(item.amountMinor, item.currency)}</td><td>{money(item.allocatedMinor, item.currency)}</td><td>{money(item.balanceMinor, item.currency)}</td><td><span className={`statusPill status-${item.state}`}>{item.state}</span></td></tr>)}</tbody></table></div> : <div className="emptyState"><h3>No rent obligations</h3><p>Rent obligations for your linked leases will appear here.</p></div>}
      </section>

      <section className="panel portalPanel">
        <div className="panelHeader"><div><p className="panelKicker">MAINTENANCE</p><h2>Your requests</h2></div><span className="countBadge">{summary.maintenance.length}</span></div>
        {summary.maintenance.length ? <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Issue</th><th>Home</th><th>Priority</th><th>Status</th><th>Updated</th></tr></thead><tbody>{summary.maintenance.map((item) => <tr key={item.id}><td><strong>{item.title}</strong><span className="recordMeta capitalize">{item.category}</span></td><td>{item.propertyName}<span className="recordMeta">{item.unitLabel}</span></td><td className="capitalize">{item.priority}</td><td><span className={`statusPill status-${item.status}`}>{item.status}</span></td><td>{new Date(item.updatedAt).toLocaleDateString("en-LK")}</td></tr>)}</tbody></table></div> : <div className="emptyState"><h3>No maintenance requests</h3><p>Requests created from this portal will appear here.</p></div>}
      </section>
    </PortalShell>
  );
}
