import { AppShell } from "../components/app-shell";
import {
  listOwners,
  listOwnershipInterests,
  listProperties,
  PropertyOSConfigurationError,
  type Owner,
  type OwnershipInterest,
  type Property,
} from "../lib/property-os";
import { createOwnerAction, createOwnershipInterestAction } from "./actions";

export const dynamic = "force-dynamic";

export default async function OwnersPage() {
  let owners: Owner[] = [];
  let interests: OwnershipInterest[] = [];
  let properties: Property[] = [];
  let errorMessage = "";

  try {
    [owners, interests, properties] = await Promise.all([
      listOwners(),
      listOwnershipInterests(),
      listProperties(),
    ]);
  } catch (error) {
    errorMessage =
      error instanceof PropertyOSConfigurationError
        ? error.message
        : "Owner data could not be loaded. Confirm migrations, PostgreSQL, and the Go API are running.";
  }

  const currentInterests = interests.filter((interest) => !interest.effectiveTo);
  const ownedProperties = new Set(currentInterests.map((interest) => interest.propertyId)).size;
  const ownershipBps = currentInterests.reduce((sum, interest) => sum + interest.ownershipBps, 0);

  return (
    <AppShell section="Owners">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">OWNERSHIP / PORTFOLIO</p>
          <h1 className="pageTitle">Owners</h1>
          <p className="pageIntro">
            Keep property ownership separate from application users, with effective ownership interests that can later drive owner statements and distributions.
          </p>
        </div>
        <div className="metricStrip" aria-label="Owner metrics">
          <div><strong>{owners.length}</strong><span>Owners</span></div>
          <div><strong>{ownedProperties}</strong><span>Owned assets</span></div>
          <div><strong>{(ownershipBps / 100).toFixed(0)}%</strong><span>Assigned interests</span></div>
        </div>
      </header>

      {errorMessage ? (
        <div className="noticePanel"><strong>Owner data is unavailable</strong><p>{errorMessage}</p></div>
      ) : null}

      <section className="leasingGrid">
        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">OWNER REGISTER</p><h2>Add owner</h2></div></header>
          <form action={createOwnerAction} className="stackedForm">
            <label>Legal / display name<input name="legalName" required placeholder="Avara Capital Holdings" /></label>
            <label>
              Owner type
              <select name="ownerType" defaultValue="individual">
                <option value="individual">Individual</option>
                <option value="company">Company</option>
              </select>
            </label>
            <div className="formPair">
              <label>Email<input name="email" type="email" /></label>
              <label>Phone<input name="phone" /></label>
            </div>
            <button className="primaryButton" type="submit">Add owner</button>
          </form>
        </article>

        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">PROPERTY INTEREST</p><h2>Assign ownership</h2></div></header>
          <form action={createOwnershipInterestAction} className="stackedForm">
            <label>
              Owner
              <select name="ownerId" required defaultValue="">
                <option value="" disabled>Select owner</option>
                {owners.filter((owner) => owner.status === "active").map((owner) => (
                  <option value={owner.id} key={owner.id}>{owner.legalName}</option>
                ))}
              </select>
            </label>
            <label>
              Property
              <select name="propertyId" required defaultValue="">
                <option value="" disabled>Select property</option>
                {properties.filter((property) => property.status === "active").map((property) => (
                  <option value={property.id} key={property.id}>{property.name}</option>
                ))}
              </select>
            </label>
            <div className="formPair">
              <label>Ownership %<input name="ownershipPercent" inputMode="decimal" required placeholder="100.00" /></label>
              <label>Effective from<input name="effectiveFrom" type="date" required /></label>
            </div>
            <button className="primaryButton" type="submit" disabled={owners.length === 0 || properties.length === 0}>Assign interest</button>
          </form>
        </article>
      </section>

      <section className="panel leasingRegister">
        <header className="panelHeader"><div><p className="panelKicker">CURRENT OWNERSHIP</p><h2>Property interests</h2></div><span className="countBadge">{currentInterests.length}</span></header>
        {currentInterests.length === 0 ? (
          <div className="emptyState"><h3>No ownership interests yet</h3><p>Add an owner and assign an effective percentage to a property.</p></div>
        ) : (
          <div className="dataTableWrap">
            <table className="dataTable">
              <thead><tr><th>Property</th><th>Owner</th><th>Interest</th><th>Effective</th></tr></thead>
              <tbody>
                {currentInterests.map((interest) => (
                  <tr key={interest.id}>
                    <td><strong>{interest.propertyName}</strong></td>
                    <td>{interest.ownerName}</td>
                    <td>{(interest.ownershipBps / 100).toFixed(2)}%</td>
                    <td>{interest.effectiveFrom}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </AppShell>
  );
}
