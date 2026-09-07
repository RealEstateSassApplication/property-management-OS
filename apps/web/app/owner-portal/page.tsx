import { PortalShell } from "../components/portal-shell";
import { getOwnerPortalSummary, PortalConfigurationError, type OwnerPortalSummary } from "../lib/portals";

export const dynamic = "force-dynamic";

function money(amountMinor: number, currency: string) {
  return new Intl.NumberFormat("en-LK", { style: "currency", currency, maximumFractionDigits: 2 }).format(amountMinor / 100);
}

export default async function OwnerPortalPage() {
  let summary: OwnerPortalSummary = { owners: [], properties: [] };
  let errorMessage = "";
  try {
    summary = await getOwnerPortalSummary();
  } catch (error) {
    errorMessage = error instanceof PortalConfigurationError ? error.message : "Owner portal data could not be loaded.";
  }

  const totalUnits = summary.properties.reduce((sum, item) => sum + item.unitCount, 0);
  const occupiedUnits = summary.properties.reduce((sum, item) => sum + item.occupiedUnits, 0);
  const openMaintenance = summary.properties.reduce((sum, item) => sum + item.openMaintenance, 0);
  const outstanding = new Map<string, number>();
  for (const property of summary.properties) {
    for (const [currency, amount] of Object.entries(property.outstandingByCurrencyMinor)) {
      outstanding.set(currency, (outstanding.get(currency) ?? 0) + amount);
    }
  }

  return (
    <PortalShell mode="owner">
      <section className="portalHero">
        <div>
          <p className="eyebrow">OWNER VIEW</p>
          <h1 className="pageTitle">Your portfolio, without the operational noise.</h1>
          <p className="pageIntro">Ownership, occupancy, receivables and maintenance exposure are scoped only to properties linked to your owner record.</p>
        </div>
      </section>

      {errorMessage ? <div className="noticePanel"><strong>Portal unavailable</strong><p>{errorMessage}</p></div> : null}

      <section className="metricStrip portalMetrics">
        <div><strong>{summary.properties.length}</strong><span>Properties</span></div>
        <div><strong>{totalUnits}</strong><span>Units</span></div>
        <div><strong>{totalUnits ? Math.round((occupiedUnits / totalUnits) * 100) : 0}%</strong><span>Occupancy</span></div>
        <div><strong>{openMaintenance}</strong><span>Open maintenance</span></div>
      </section>

      <section className="panel portalPanel">
        <div className="panelHeader"><div><p className="panelKicker">PORTFOLIO</p><h2>Linked properties</h2></div><span className="countBadge">{summary.properties.length}</span></div>
        {summary.properties.length ? (
          <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Property</th><th>Ownership</th><th>Occupancy</th><th>Maintenance</th><th>Outstanding</th></tr></thead><tbody>
            {summary.properties.map((property) => (
              <tr key={`${property.ownerId}:${property.propertyId}`}>
                <td><strong>{property.name}</strong><span className="recordMeta">{property.referenceCode} · {property.city}</span></td>
                <td>{(property.ownershipBps / 100).toFixed(2)}%</td>
                <td>{property.occupiedUnits}/{property.unitCount}</td>
                <td>{property.openMaintenance}</td>
                <td>{Object.entries(property.outstandingByCurrencyMinor).length ? Object.entries(property.outstandingByCurrencyMinor).map(([currency, amount]) => <span className="recordMeta" key={currency}>{money(amount, currency)}</span>) : "—"}</td>
              </tr>
            ))}
          </tbody></table></div>
        ) : <div className="emptyState"><h3>No linked properties</h3><p>Your manager must explicitly link your login to an owner record before portfolio data appears.</p></div>}
      </section>

      {outstanding.size ? <section className="portalSummaryLine">{Array.from(outstanding.entries()).map(([currency, amount]) => <span key={currency}><strong>{money(amount, currency)}</strong> outstanding across linked properties</span>)}</section> : null}
    </PortalShell>
  );
}
