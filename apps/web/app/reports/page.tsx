import { AppShell } from "../components/app-shell";
import {
  AdminConfigurationError,
  getReportingDashboard,
  listAuditEvents,
  type AuditEvent,
  type CurrencyAmount,
  type ReportingDashboard,
} from "../lib/admin";

export const dynamic = "force-dynamic";

function money(item: CurrencyAmount) {
  return new Intl.NumberFormat("en-LK", { style: "currency", currency: item.currency, maximumFractionDigits: 2 }).format(item.amountMinor / 100);
}

function currencyLines(items: CurrencyAmount[]) {
  if (!items.length) return <span className="recordMeta">None</span>;
  return items.map((item) => <span className="reportMoney" key={item.currency}>{money(item)}</span>);
}

export default async function ReportsPage() {
  let dashboard: ReportingDashboard = {
    propertyCount: 0, unitCount: 0, occupiedUnits: 0, vacancyRateBps: 0,
    openMaintenance: 0, emergencyMaintenance: 0, leasesExpiring30Days: 0, leasesExpiring90Days: 0,
    outstandingByCurrency: [], overdueByCurrency: [], collectedThisMonthByCurrency: [],
  };
  let audit: AuditEvent[] = [];
  let errorMessage = "";
  try {
    [dashboard, audit] = await Promise.all([getReportingDashboard(), listAuditEvents(100)]);
  } catch (error) {
    errorMessage = error instanceof AdminConfigurationError ? error.message : "Reporting data could not be loaded. Confirm the Go API and latest migrations are running.";
  }

  const occupancyRate = dashboard.unitCount ? Math.round((dashboard.occupiedUnits / dashboard.unitCount) * 100) : 0;

  return (
    <AppShell section="Reports">
      <header className="workspaceHeader">
        <div><p className="eyebrow">OPERATING REVIEW</p><h1 className="pageTitle">Portfolio reporting</h1><p className="pageIntro">Live operational and financial indicators derived from the portfolio, lease, rent and maintenance source-of-truth tables.</p></div>
        <div className="metricStrip"><div><strong>{occupancyRate}%</strong><span>Occupancy</span></div><div><strong>{(dashboard.vacancyRateBps / 100).toFixed(1)}%</strong><span>Vacancy</span></div><div><strong>{dashboard.openMaintenance}</strong><span>Open work</span></div></div>
      </header>

      {errorMessage ? <div className="noticePanel"><strong>Reporting unavailable</strong><p>{errorMessage}</p></div> : null}

      <section className="reportMetricGrid">
        <article className="reportCard"><span>Properties</span><strong>{dashboard.propertyCount}</strong><small>{dashboard.unitCount} total units</small></article>
        <article className="reportCard"><span>Occupied units</span><strong>{dashboard.occupiedUnits}</strong><small>{dashboard.unitCount - dashboard.occupiedUnits} not occupied</small></article>
        <article className="reportCard"><span>Lease expiry</span><strong>{dashboard.leasesExpiring30Days}</strong><small>{dashboard.leasesExpiring90Days} within 90 days</small></article>
        <article className="reportCard"><span>Emergency maintenance</span><strong>{dashboard.emergencyMaintenance}</strong><small>{dashboard.openMaintenance} open total</small></article>
      </section>

      <div className="leasingGrid">
        <section className="panel reportFinancialPanel"><div className="panelHeader"><div><p className="panelKicker">RECEIVABLES</p><h2>Outstanding</h2></div></div><div className="reportMoneyList">{currencyLines(dashboard.outstandingByCurrency)}</div></section>
        <section className="panel reportFinancialPanel"><div className="panelHeader"><div><p className="panelKicker">ARREARS</p><h2>Overdue</h2></div></div><div className="reportMoneyList">{currencyLines(dashboard.overdueByCurrency)}</div></section>
      </div>
      <section className="panel leasingRegister reportFinancialPanel"><div className="panelHeader"><div><p className="panelKicker">CASH</p><h2>Collected this month</h2></div></div><div className="reportMoneyList">{currencyLines(dashboard.collectedThisMonthByCurrency)}</div></section>

      <section className="panel leasingRegister">
        <div className="panelHeader"><div><p className="panelKicker">AUDIT TRAIL</p><h2>Recent system activity</h2></div><span className="countBadge">{audit.length}</span></div>
        {audit.length ? <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Time</th><th>Actor</th><th>Action</th><th>Resource</th><th>Context</th></tr></thead><tbody>{audit.map((event) => <tr key={event.id}><td>{new Date(event.occurredAt).toLocaleString("en-LK")}</td><td>{event.actorName || "System"}<span className="recordMeta">{event.actorUserId || "—"}</span></td><td><strong>{event.action}</strong></td><td>{event.resourceType}<span className="recordMeta">{event.resourceId || "—"}</span></td><td><code className="auditMeta">{Object.keys(event.metadata ?? {}).length ? JSON.stringify(event.metadata) : "{}"}</code></td></tr>)}</tbody></table></div> : <div className="emptyState"><h3>No audit events</h3><p>Domain actions that write audit events will appear here as the organization operates.</p></div>}
      </section>
    </AppShell>
  );
}
