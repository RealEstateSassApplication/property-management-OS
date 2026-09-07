import { AppShell } from "../components/app-shell";
import {
  listLeases,
  listProperties,
  listTenancies,
  listTenants,
  listUnits,
  PropertyOSConfigurationError,
  type Lease,
  type Property,
  type Tenancy,
  type Tenant,
  type Unit,
} from "../lib/property-os";
import { createLeaseAction, createTenancyAction } from "./actions";

export const dynamic = "force-dynamic";

type UnitOption = Unit & { propertyName: string };

function formatMoney(amountMinor: number, currency: string) {
  return new Intl.NumberFormat("en-LK", {
    style: "currency",
    currency,
    maximumFractionDigits: 2,
  }).format(amountMinor / 100);
}

export default async function LeasingPage() {
  let properties: Property[] = [];
  let tenants: Tenant[] = [];
  let tenancies: Tenancy[] = [];
  let leases: Lease[] = [];
  let units: UnitOption[] = [];
  let errorMessage = "";

  try {
    [properties, tenants, tenancies, leases] = await Promise.all([
      listProperties(),
      listTenants(),
      listTenancies(),
      listLeases(),
    ]);
    const unitGroups = await Promise.all(
      properties.map(async (property) =>
        (await listUnits(property.id)).map((unit) => ({ ...unit, propertyName: property.name })),
      ),
    );
    units = unitGroups.flat();
  } catch (error) {
    errorMessage =
      error instanceof PropertyOSConfigurationError
        ? error.message
        : "Leasing data could not be loaded. Confirm migrations, PostgreSQL, and the Go API are running.";
  }

  const activeTenancies = tenancies.filter((tenancy) => tenancy.status === "active").length;
  const activeLeases = leases.filter((lease) => lease.status === "active").length;
  const availableUnits = units.filter((unit) => unit.occupancyStatus === "vacant" || unit.occupancyStatus === "reserved");
  const eligibleTenants = tenants.filter((tenant) => tenant.status !== "blocked" && tenant.status !== "former");
  const leaseableTenancies = tenancies.filter((tenancy) => tenancy.status === "active" || tenancy.status === "upcoming");

  return (
    <AppShell section="Leasing">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">OCCUPANCY / CONTRACTS</p>
          <h1 className="pageTitle">Leasing</h1>
          <p className="pageIntro">
            Move a renter from a unit assignment into a controlled lease lifecycle without mixing occupancy state and financial terms.
          </p>
        </div>
        <div className="metricStrip" aria-label="Leasing metrics">
          <div><strong>{activeTenancies}</strong><span>Occupied</span></div>
          <div><strong>{activeLeases}</strong><span>Active leases</span></div>
          <div><strong>{availableUnits.length}</strong><span>Available units</span></div>
        </div>
      </header>

      {errorMessage ? (
        <div className="noticePanel">
          <strong>Leasing data is unavailable</strong>
          <p>{errorMessage}</p>
        </div>
      ) : null}

      <section className="leasingGrid">
        <article className="panel">
          <header className="panelHeader">
            <div><p className="panelKicker">STEP 01</p><h2>Create tenancy</h2></div>
          </header>
          <form action={createTenancyAction} className="stackedForm">
            <label>
              Unit
              <select name="unitId" required defaultValue="">
                <option value="" disabled>Select an available unit</option>
                {availableUnits.map((unit) => (
                  <option value={unit.id} key={unit.id}>{unit.propertyName} · {unit.label}</option>
                ))}
              </select>
            </label>
            <label>
              Primary tenant
              <select name="primaryTenantId" required defaultValue="">
                <option value="" disabled>Select a tenant</option>
                {eligibleTenants.map((tenant) => (
                  <option value={tenant.id} key={tenant.id}>{tenant.legalName}</option>
                ))}
              </select>
            </label>
            <div className="formPair">
              <label>Start date<input name="startDate" type="date" required /></label>
              <label>End date<input name="endDate" type="date" /></label>
            </div>
            <label>
              Occupancy state
              <select name="status" defaultValue="upcoming">
                <option value="upcoming">Upcoming</option>
                <option value="active">Active now</option>
              </select>
            </label>
            <button className="primaryButton" type="submit" disabled={availableUnits.length === 0 || eligibleTenants.length === 0}>
              Create tenancy
            </button>
          </form>
        </article>

        <article className="panel">
          <header className="panelHeader">
            <div><p className="panelKicker">STEP 02</p><h2>Create lease</h2></div>
          </header>
          <form action={createLeaseAction} className="stackedForm">
            <label>
              Tenancy
              <select name="tenancyId" required defaultValue="">
                <option value="" disabled>Select occupancy</option>
                {leaseableTenancies.map((tenancy) => (
                  <option value={tenancy.id} key={tenancy.id}>{tenancy.primaryTenantName} · {tenancy.propertyName} / {tenancy.unitLabel}</option>
                ))}
              </select>
            </label>
            <label>Reference code<input name="referenceCode" required placeholder="LEASE-COL-002-2026" /></label>
            <div className="formPair">
              <label>Start date<input name="startDate" type="date" required /></label>
              <label>End date<input name="endDate" type="date" required /></label>
            </div>
            <div className="formPair">
              <label>Monthly rent (LKR)<input name="rentAmount" inputMode="decimal" required placeholder="150000.00" /></label>
              <label>Deposit (LKR)<input name="depositAmount" inputMode="decimal" defaultValue="0.00" /></label>
            </div>
            <div className="formPair">
              <label>Due day<input name="dueDay" type="number" min="1" max="31" defaultValue="1" required /></label>
              <label>
                Lease state
                <select name="status" defaultValue="draft"><option value="draft">Draft</option><option value="active">Activate now</option></select>
              </label>
            </div>
            <input name="currency" type="hidden" value="LKR" />
            <button className="primaryButton" type="submit" disabled={leaseableTenancies.length === 0}>Create lease</button>
          </form>
        </article>
      </section>

      <section className="panel leasingRegister">
        <header className="panelHeader">
          <div><p className="panelKicker">CONTRACT REGISTER</p><h2>Leases</h2></div>
          <span className="countBadge">{leases.length}</span>
        </header>
        {leases.length === 0 ? (
          <div className="emptyState"><h3>No leases yet</h3><p>Create a tenancy first, then capture the contract terms here.</p></div>
        ) : (
          <div className="dataTableWrap">
            <table className="dataTable">
              <thead><tr><th>Lease</th><th>Resident</th><th>Unit</th><th>Term</th><th>Rent</th><th>Status</th></tr></thead>
              <tbody>
                {leases.map((lease) => (
                  <tr key={lease.id}>
                    <td><strong>{lease.referenceCode}</strong><span className="recordMeta">Due day {lease.dueDay}</span></td>
                    <td>{lease.primaryTenantName}</td>
                    <td>{lease.propertyName}<span className="recordMeta">{lease.unitLabel}</span></td>
                    <td>{lease.startDate}<span className="recordMeta">to {lease.endDate}</span></td>
                    <td>{formatMoney(lease.rentAmountMinor, lease.currency)}<span className="recordMeta">Deposit {formatMoney(lease.depositAmountMinor, lease.currency)}</span></td>
                    <td><span className={`statusPill status-${lease.status}`}>{lease.status}</span></td>
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
