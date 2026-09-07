import { AppShell } from "../components/app-shell";
import {
  listLeases,
  listRentObligations,
  listRentPayments,
  listTenants,
  PropertyOSConfigurationError,
  type Lease,
  type RentObligation,
  type RentPayment,
  type Tenant,
} from "../lib/property-os";
import { createRentAllocationAction, createRentObligationAction, createRentPaymentAction } from "./actions";

export const dynamic = "force-dynamic";

function formatMoney(amountMinor: number, currency: string) {
  return new Intl.NumberFormat("en-LK", {
    style: "currency",
    currency,
    maximumFractionDigits: 2,
  }).format(amountMinor / 100);
}

export default async function RentPage() {
  let leases: Lease[] = [];
  let tenants: Tenant[] = [];
  let obligations: RentObligation[] = [];
  let payments: RentPayment[] = [];
  let errorMessage = "";

  try {
    [leases, tenants, obligations, payments] = await Promise.all([
      listLeases(),
      listTenants(),
      listRentObligations(),
      listRentPayments(),
    ]);
  } catch (error) {
    errorMessage =
      error instanceof PropertyOSConfigurationError
        ? error.message
        : "Rent data could not be loaded. Confirm migration 000003, PostgreSQL, and the Go API are running.";
  }

  const activeLeases = leases.filter((lease) => lease.status === "active");
  const payableObligations = obligations.filter((obligation) => obligation.state !== "paid" && obligation.state !== "void" && obligation.balanceMinor > 0);
  const availablePayments = payments.filter((payment) => payment.status === "posted" && payment.unallocatedMinor > 0);
  const outstandingLkr = obligations.filter((obligation) => obligation.currency === "LKR" && obligation.state !== "void").reduce((sum, item) => sum + item.balanceMinor, 0);
  const overdueCount = obligations.filter((obligation) => obligation.state === "overdue").length;
  const unallocatedLkr = payments.filter((payment) => payment.currency === "LKR" && payment.status === "posted").reduce((sum, item) => sum + item.unallocatedMinor, 0);
  const today = new Date().toISOString().slice(0, 10);

  return (
    <AppShell section="Rent">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">RENT / RECEIVABLES</p>
          <h1 className="pageTitle">Rent</h1>
          <p className="pageIntro">
            Generate lease-backed obligations, record immutable payments, and allocate cash without hiding arrears behind a paid/unpaid flag.
          </p>
        </div>
        <div className="metricStrip" aria-label="Rent metrics">
          <div><strong>{formatMoney(outstandingLkr, "LKR")}</strong><span>Outstanding</span></div>
          <div><strong>{overdueCount}</strong><span>Overdue</span></div>
          <div><strong>{formatMoney(unallocatedLkr, "LKR")}</strong><span>Unallocated cash</span></div>
        </div>
      </header>

      {errorMessage ? (
        <div className="noticePanel"><strong>Rent data is unavailable</strong><p>{errorMessage}</p></div>
      ) : null}

      <section className="leasingGrid">
        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">STEP 01</p><h2>Create obligation</h2></div></header>
          <form action={createRentObligationAction} className="stackedForm">
            <label>
              Active lease
              <select name="leaseId" required defaultValue="">
                <option value="" disabled>Select lease</option>
                {activeLeases.map((lease) => (
                  <option value={lease.id} key={lease.id}>{lease.referenceCode} · {lease.primaryTenantName} · {formatMoney(lease.rentAmountMinor, lease.currency)}</option>
                ))}
              </select>
            </label>
            <label>Rent month<input name="period" type="month" required /></label>
            <p className="recordMeta">Amount, currency and due day come from the lease on the Go backend.</p>
            <button className="primaryButton" type="submit" disabled={activeLeases.length === 0}>Generate obligation</button>
          </form>
        </article>

        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">STEP 02</p><h2>Record payment</h2></div></header>
          <form action={createRentPaymentAction} className="stackedForm">
            <label>
              Tenant
              <select name="tenantId" required defaultValue="">
                <option value="" disabled>Select tenant</option>
                {tenants.filter((tenant) => tenant.status === "active").map((tenant) => (
                  <option value={tenant.id} key={tenant.id}>{tenant.legalName}</option>
                ))}
              </select>
            </label>
            <div className="formPair">
              <label>Amount (LKR)<input name="amount" inputMode="decimal" required placeholder="150000.00" /></label>
              <label>Received<input name="receivedAt" type="date" required defaultValue={today} /></label>
            </div>
            <div className="formPair">
              <label>
                Method
                <select name="method" defaultValue="bank_transfer">
                  <option value="bank_transfer">Bank transfer</option>
                  <option value="cash">Cash</option>
                  <option value="card">Card</option>
                  <option value="online">Online</option>
                  <option value="other">Other</option>
                </select>
              </label>
              <label>Reference<input name="referenceCode" placeholder="BANK-REF-001" /></label>
            </div>
            <input name="currency" type="hidden" value="LKR" />
            <button className="primaryButton" type="submit" disabled={tenants.length === 0}>Record payment</button>
          </form>
        </article>

        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">STEP 03</p><h2>Allocate payment</h2></div></header>
          <form action={createRentAllocationAction} className="stackedForm">
            <label>
              Available payment
              <select name="paymentId" required defaultValue="">
                <option value="" disabled>Select payment</option>
                {availablePayments.map((payment) => (
                  <option value={payment.id} key={payment.id}>{payment.tenantName} · {formatMoney(payment.unallocatedMinor, payment.currency)} available</option>
                ))}
              </select>
            </label>
            <label>
              Outstanding obligation
              <select name="obligationId" required defaultValue="">
                <option value="" disabled>Select obligation</option>
                {payableObligations.map((obligation) => (
                  <option value={obligation.id} key={obligation.id}>{obligation.primaryTenantName} · {obligation.period} · {formatMoney(obligation.balanceMinor, obligation.currency)}</option>
                ))}
              </select>
            </label>
            <label>Allocate amount<input name="amount" inputMode="decimal" required placeholder="100000.00" /></label>
            <p className="recordMeta">The API rejects wrong-tenant, wrong-currency, and over-allocation attempts under row locks.</p>
            <button className="primaryButton" type="submit" disabled={availablePayments.length === 0 || payableObligations.length === 0}>Allocate cash</button>
          </form>
        </article>
      </section>

      <section className="panel leasingRegister">
        <header className="panelHeader"><div><p className="panelKicker">RECEIVABLES</p><h2>Rent obligations</h2></div><span className="countBadge">{obligations.length}</span></header>
        {obligations.length === 0 ? (
          <div className="emptyState"><h3>No obligations yet</h3><p>Generate the first monthly obligation from an active lease.</p></div>
        ) : (
          <div className="dataTableWrap">
            <table className="dataTable">
              <thead><tr><th>Period</th><th>Resident</th><th>Asset</th><th>Due</th><th>Amount</th><th>Balance</th><th>State</th></tr></thead>
              <tbody>{obligations.map((item) => (
                <tr key={item.id}>
                  <td><strong>{item.period}</strong><span className="recordMeta">{item.leaseReference}</span></td>
                  <td>{item.primaryTenantName}</td>
                  <td>{item.propertyName}<span className="recordMeta">{item.unitLabel}</span></td>
                  <td>{item.dueDate}</td>
                  <td>{formatMoney(item.amountMinor, item.currency)}</td>
                  <td>{formatMoney(item.balanceMinor, item.currency)}<span className="recordMeta">Allocated {formatMoney(item.allocatedMinor, item.currency)}</span></td>
                  <td><span className={`statusPill status-${item.state}`}>{item.state}</span></td>
                </tr>
              ))}</tbody>
            </table>
          </div>
        )}
      </section>

      <section className="panel leasingRegister">
        <header className="panelHeader"><div><p className="panelKicker">CASH REGISTER</p><h2>Payments</h2></div><span className="countBadge">{payments.length}</span></header>
        {payments.length === 0 ? (
          <div className="emptyState"><h3>No payments yet</h3><p>Record a tenant payment, then allocate it to an outstanding obligation.</p></div>
        ) : (
          <div className="dataTableWrap">
            <table className="dataTable">
              <thead><tr><th>Received</th><th>Tenant</th><th>Reference</th><th>Amount</th><th>Allocated</th><th>Available</th></tr></thead>
              <tbody>{payments.map((payment) => (
                <tr key={payment.id}>
                  <td>{payment.receivedAt}</td>
                  <td><strong>{payment.tenantName}</strong><span className="recordMeta">{payment.method.replaceAll("_", " ")}</span></td>
                  <td>{payment.referenceCode || "—"}</td>
                  <td>{formatMoney(payment.amountMinor, payment.currency)}</td>
                  <td>{formatMoney(payment.allocatedMinor, payment.currency)}</td>
                  <td>{formatMoney(payment.unallocatedMinor, payment.currency)}</td>
                </tr>
              ))}</tbody>
            </table>
          </div>
        )}
      </section>
    </AppShell>
  );
}
