import Link from "next/link";
import { AppShell } from "../components/app-shell";
import { listProperties, type Property } from "../lib/property-os";
import { createPropertyAction } from "./actions";

export const dynamic = "force-dynamic";

export default async function PropertiesPage() {
  let properties: Property[] = [];
  let loadError: string | null = null;

  try {
    properties = await listProperties();
  } catch (error) {
    loadError = error instanceof Error ? error.message : "Could not load the portfolio.";
  }

  const activeCount = properties.filter((property) => property.status === "active").length;

  return (
    <AppShell section="Portfolio">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">PORTFOLIO</p>
          <h1 className="pageTitle">Managed properties</h1>
          <p className="pageIntro">The source of truth for buildings, houses, commercial sites, and their units.</p>
        </div>
        <div className="metricStrip">
          <div><strong>{properties.length}</strong><span>Total properties</span></div>
          <div><strong>{activeCount}</strong><span>Active</span></div>
        </div>
      </header>

      {loadError ? (
        <section className="noticePanel">
          <strong>Portfolio API is not connected.</strong>
          <p>{loadError}</p>
          <code>make db-up && make migrate-up && make seed && make api</code>
        </section>
      ) : null}

      <section className="contentGrid">
        <div className="panel panelWide">
          <div className="panelHeader">
            <div>
              <p className="panelKicker">PORTFOLIO REGISTER</p>
              <h2>Properties</h2>
            </div>
            <span className="countBadge">{properties.length}</span>
          </div>

          {properties.length === 0 ? (
            <div className="emptyState">
              <h3>No properties yet</h3>
              <p>Add the first managed property using the form beside the register.</p>
            </div>
          ) : (
            <div className="dataTableWrap">
              <table className="dataTable">
                <thead>
                  <tr>
                    <th>Property</th>
                    <th>Type</th>
                    <th>Location</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {properties.map((property) => (
                    <tr key={property.id}>
                      <td>
                        <Link className="recordLink" href={`/properties/${property.id}`}>{property.name}</Link>
                        <span className="recordMeta">{property.referenceCode ?? "No reference"}</span>
                      </td>
                      <td className="capitalize">{property.propertyType}</td>
                      <td>{property.city}, {property.countryCode}</td>
                      <td><span className={`statusPill status-${property.status}`}>{property.status}</span></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        <aside className="panel">
          <div className="panelHeader">
            <div>
              <p className="panelKicker">QUICK ENTRY</p>
              <h2>Add property</h2>
            </div>
          </div>
          <form action={createPropertyAction} className="stackedForm">
            <label>Name<input name="name" placeholder="Park Residences" required /></label>
            <div className="formPair">
              <label>Reference<input name="referenceCode" placeholder="COL-002" /></label>
              <label>Type
                <select name="propertyType" defaultValue="building">
                  <option value="building">Building</option>
                  <option value="apartment">Apartment</option>
                  <option value="house">House</option>
                  <option value="commercial">Commercial</option>
                  <option value="land">Land</option>
                  <option value="other">Other</option>
                </select>
              </label>
            </div>
            <label>Address<input name="addressLine1" placeholder="18 Park Road" required /></label>
            <div className="formPair">
              <label>City<input name="city" placeholder="Colombo 05" required /></label>
              <label>Region<input name="region" placeholder="Western" /></label>
            </div>
            <label>Country code<input name="countryCode" defaultValue="LK" maxLength={2} required /></label>
            <button className="primaryButton" type="submit">Create property</button>
          </form>
        </aside>
      </section>
    </AppShell>
  );
}
