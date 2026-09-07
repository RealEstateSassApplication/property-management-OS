import Link from "next/link";
import { AppShell } from "../../components/app-shell";
import { getProperty, listUnits, type Property, type Unit } from "../../lib/property-os";
import { createUnitAction } from "../actions";

export const dynamic = "force-dynamic";

export default async function PropertyDetailPage({ params }: { params: Promise<{ propertyID: string }> }) {
  const { propertyID } = await params;
  let property: Property | null = null;
  let units: Unit[] = [];
  let loadError: string | null = null;

  try {
    [property, units] = await Promise.all([getProperty(propertyID), listUnits(propertyID)]);
  } catch (error) {
    loadError = error instanceof Error ? error.message : "Could not load property.";
  }

  if (!property) {
    return (
      <AppShell section="Portfolio">
        <Link className="backLink" href="/properties">← Portfolio</Link>
        <section className="noticePanel">
          <strong>Property could not be loaded.</strong>
          <p>{loadError}</p>
        </section>
      </AppShell>
    );
  }

  const occupied = units.filter((unit) => unit.occupancyStatus === "occupied").length;
  const occupancyRate = units.length === 0 ? 0 : Math.round((occupied / units.length) * 100);

  return (
    <AppShell section="Portfolio">
      <Link className="backLink" href="/properties">← Portfolio</Link>
      <header className="workspaceHeader detailHeader">
        <div>
          <p className="eyebrow">{property.referenceCode ?? "PROPERTY"}</p>
          <h1 className="pageTitle">{property.name}</h1>
          <p className="pageIntro">{property.addressLine1}, {property.city}{property.region ? `, ${property.region}` : ""}</p>
        </div>
        <div className="metricStrip">
          <div><strong>{units.length}</strong><span>Units</span></div>
          <div><strong>{occupied}</strong><span>Occupied</span></div>
          <div><strong>{occupancyRate}%</strong><span>Occupancy</span></div>
        </div>
      </header>

      <section className="propertyFacts">
        <div><span>Type</span><strong className="capitalize">{property.propertyType}</strong></div>
        <div><span>Status</span><strong className="capitalize">{property.status}</strong></div>
        <div><span>Country</span><strong>{property.countryCode}</strong></div>
        <div><span>Avara link</span><strong>{property.externalAvaraPropertyId ? "Linked" : "Not linked"}</strong></div>
      </section>

      <section className="contentGrid">
        <div className="panel panelWide">
          <div className="panelHeader">
            <div>
              <p className="panelKicker">UNIT REGISTER</p>
              <h2>Units & occupancy</h2>
            </div>
            <span className="countBadge">{units.length}</span>
          </div>
          {units.length === 0 ? (
            <div className="emptyState"><h3>No units yet</h3><p>Create the first rentable or occupiable unit for this property.</p></div>
          ) : (
            <div className="dataTableWrap">
              <table className="dataTable">
                <thead><tr><th>Unit</th><th>Configuration</th><th>Area</th><th>Occupancy</th></tr></thead>
                <tbody>
                  {units.map((unit) => (
                    <tr key={unit.id}>
                      <td><strong>{unit.label}</strong><span className="recordMeta">{unit.referenceCode}</span></td>
                      <td>{unit.bedrooms ?? "—"} bd · {unit.bathrooms ?? "—"} ba</td>
                      <td>{unit.floorArea ? `${unit.floorArea.toLocaleString()} ${unit.floorAreaUnit ?? ""}` : "—"}</td>
                      <td><span className={`statusPill status-${unit.occupancyStatus}`}>{unit.occupancyStatus}</span></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        <aside className="panel">
          <div className="panelHeader">
            <div><p className="panelKicker">QUICK ENTRY</p><h2>Add unit</h2></div>
          </div>
          <form action={createUnitAction} className="stackedForm">
            <input type="hidden" name="propertyId" value={property.id} />
            <div className="formPair">
              <label>Reference<input name="referenceCode" placeholder="A-03" required /></label>
              <label>Label<input name="label" placeholder="Apartment 03" required /></label>
            </div>
            <div className="formPair">
              <label>Bedrooms<input name="bedrooms" min="0" step="0.5" type="number" /></label>
              <label>Bathrooms<input name="bathrooms" min="0" step="0.5" type="number" /></label>
            </div>
            <div className="formPair">
              <label>Floor area<input name="floorArea" min="0" step="0.01" type="number" /></label>
              <label>Unit<select name="floorAreaUnit" defaultValue="sqft"><option value="sqft">sq ft</option><option value="sqm">sq m</option></select></label>
            </div>
            <button className="primaryButton" type="submit">Add unit</button>
          </form>
        </aside>
      </section>
    </AppShell>
  );
}
