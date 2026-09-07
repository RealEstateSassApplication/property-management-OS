import { AppShell } from "../components/app-shell";
import {
  listMaintenanceEvidence,
  listMaintenanceQuotes,
  listMaintenanceRequests,
  listMaintenanceVendors,
  listMaintenanceWorkOrders,
  listProperties,
  listTenants,
  listUnits,
  PropertyOSConfigurationError,
  type MaintenanceEvidence,
  type MaintenanceQuote,
  type MaintenanceRequest,
  type MaintenanceVendor,
  type MaintenanceWorkOrder,
  type Property,
  type Tenant,
  type Unit,
} from "../lib/property-os";
import {
  createEvidenceAction,
  createQuoteAction,
  createRequestAction,
  createVendorAction,
  createWorkOrderAction,
  decideQuoteAction,
  updateRequestStatusAction,
  updateWorkOrderStatusAction,
} from "./actions";

export const dynamic = "force-dynamic";

type UnitOption = Unit & { propertyName: string };

function formatMoney(amountMinor: number, currency: string) {
  return new Intl.NumberFormat("en-LK", { style: "currency", currency, maximumFractionDigits: 2 }).format(amountMinor / 100);
}

export default async function MaintenancePage() {
  let properties: Property[] = [];
  let units: UnitOption[] = [];
  let tenants: Tenant[] = [];
  let vendors: MaintenanceVendor[] = [];
  let requests: MaintenanceRequest[] = [];
  let workOrders: MaintenanceWorkOrder[] = [];
  let quotes: MaintenanceQuote[] = [];
  let evidence: MaintenanceEvidence[] = [];
  let errorMessage = "";

  try {
    [properties, tenants, vendors, requests, workOrders, quotes, evidence] = await Promise.all([
      listProperties(),
      listTenants(),
      listMaintenanceVendors(),
      listMaintenanceRequests(),
      listMaintenanceWorkOrders(),
      listMaintenanceQuotes(),
      listMaintenanceEvidence(),
    ]);
    const groups = await Promise.all(
      properties.map(async (property) =>
        (await listUnits(property.id)).map((unit) => ({ ...unit, propertyName: property.name })),
      ),
    );
    units = groups.flat();
  } catch (error) {
    errorMessage =
      error instanceof PropertyOSConfigurationError
        ? error.message
        : "Maintenance data could not be loaded. Confirm migrations, PostgreSQL, and the Go API are running.";
  }

  const activeRequests = requests.filter((item) => !["resolved", "cancelled"].includes(item.status));
  const activeWorkOrders = workOrders.filter((item) => !["completed", "cancelled"].includes(item.status));
  const emergencyCount = activeRequests.filter((item) => item.priority === "emergency").length;
  const submittedQuotes = quotes.filter((item) => item.status === "submitted");
  const activeVendors = vendors.filter((item) => item.status === "active");
  const evidenceByWorkOrder = new Map<string, number>();
  for (const item of evidence) evidenceByWorkOrder.set(item.workOrderId, (evidenceByWorkOrder.get(item.workOrderId) ?? 0) + 1);

  return (
    <AppShell section="Maintenance">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">SERVICE OPERATIONS</p>
          <h1 className="pageTitle">Maintenance</h1>
          <p className="pageIntro">
            Capture issues, dispatch controlled work orders, separate quote approval from execution, and require proof before jobs can close.
          </p>
        </div>
        <div className="metricStrip" aria-label="Maintenance metrics">
          <div><strong>{activeRequests.length}</strong><span>Open requests</span></div>
          <div><strong>{emergencyCount}</strong><span>Emergency</span></div>
          <div><strong>{activeWorkOrders.length}</strong><span>Active jobs</span></div>
          <div><strong>{submittedQuotes.length}</strong><span>Quotes awaiting decision</span></div>
        </div>
      </header>

      {errorMessage ? (
        <div className="noticePanel"><strong>Maintenance data is unavailable</strong><p>{errorMessage}</p></div>
      ) : null}

      <section className="leasingGrid">
        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">INTAKE</p><h2>New maintenance request</h2></div></header>
          <form action={createRequestAction} className="stackedForm">
            <label>
              Property
              <select name="propertyId" required defaultValue="">
                <option value="" disabled>Select property</option>
                {properties.map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label>
              Unit (optional)
              <select name="unitId" defaultValue="">
                <option value="">Property-level issue</option>
                {units.map((item) => <option value={item.id} key={item.id}>{item.propertyName} · {item.label}</option>)}
              </select>
            </label>
            <label>
              Tenant (optional; must actively occupy selected unit)
              <select name="tenantId" defaultValue="">
                <option value="">No tenant link</option>
                {tenants.filter((item) => item.status === "active").map((item) => <option value={item.id} key={item.id}>{item.legalName}</option>)}
              </select>
            </label>
            <label>Issue title<input name="title" required placeholder="Kitchen sink leak" /></label>
            <label>Description<textarea name="description" required rows={3} placeholder="Describe symptoms, access constraints, and urgency." /></label>
            <div className="formPair">
              <label>
                Category
                <select name="category" defaultValue="plumbing">
                  <option value="plumbing">Plumbing</option><option value="electrical">Electrical</option><option value="hvac">HVAC</option>
                  <option value="appliance">Appliance</option><option value="structural">Structural</option><option value="cleaning">Cleaning</option>
                  <option value="security">Security</option><option value="other">Other</option>
                </select>
              </label>
              <label>
                Priority
                <select name="priority" defaultValue="normal">
                  <option value="low">Low</option><option value="normal">Normal</option><option value="high">High</option><option value="emergency">Emergency</option>
                </select>
              </label>
            </div>
            <button className="primaryButton" type="submit" disabled={properties.length === 0}>Create request</button>
          </form>
        </article>

        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">DISPATCH</p><h2>Create work order</h2></div></header>
          <form action={createWorkOrderAction} className="stackedForm">
            <label>
              Request
              <select name="maintenanceRequestId" required defaultValue="">
                <option value="" disabled>Select open request</option>
                {activeRequests.map((item) => <option value={item.id} key={item.id}>{item.priority.toUpperCase()} · {item.title}</option>)}
              </select>
            </label>
            <label>
              Vendor (optional)
              <select name="vendorId" defaultValue="">
                <option value="">Internal / assign later</option>
                {activeVendors.map((item) => <option value={item.id} key={item.id}>{item.name} · {item.trade}</option>)}
              </select>
            </label>
            <label>Scope / work summary<textarea name="summary" required rows={3} placeholder="Inspect failed fitting and replace if required." /></label>
            <label>Schedule<input name="scheduledFor" type="datetime-local" /></label>
            <button className="primaryButton" type="submit" disabled={activeRequests.length === 0}>Create work order</button>
          </form>
        </article>

        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">SUPPLIERS</p><h2>Add vendor</h2></div></header>
          <form action={createVendorAction} className="stackedForm">
            <label>Vendor name<input name="name" required placeholder="Colombo Rapid Plumbing" /></label>
            <label>
              Trade
              <select name="trade" defaultValue="general">
                <option value="general">General</option><option value="plumbing">Plumbing</option><option value="electrical">Electrical</option>
                <option value="hvac">HVAC</option><option value="appliance">Appliance</option><option value="structural">Structural</option>
                <option value="cleaning">Cleaning</option><option value="security">Security</option><option value="other">Other</option>
              </select>
            </label>
            <div className="formPair"><label>Email<input name="email" type="email" /></label><label>Phone<input name="phone" /></label></div>
            <button className="primaryButton" type="submit">Add vendor</button>
          </form>
        </article>

        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">COST CONTROL</p><h2>Submit vendor quote</h2></div></header>
          <form action={createQuoteAction} className="stackedForm">
            <label>
              Work order
              <select name="workOrderId" required defaultValue="">
                <option value="" disabled>Select active work order</option>
                {activeWorkOrders.map((item) => <option value={item.id} key={item.id}>{item.requestTitle} · {item.summary}</option>)}
              </select>
            </label>
            <label>
              Vendor
              <select name="vendorId" required defaultValue="">
                <option value="" disabled>Select vendor</option>
                {activeVendors.map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}
              </select>
            </label>
            <label>Scope quoted<textarea name="scopeSummary" required rows={3} /></label>
            <div className="formPair"><label>Amount (LKR)<input name="amount" required inputMode="decimal" placeholder="18500.00" /></label><label>Currency<input name="currency" defaultValue="LKR" maxLength={3} /></label></div>
            <button className="primaryButton" type="submit" disabled={activeWorkOrders.length === 0 || activeVendors.length === 0}>Submit quote</button>
          </form>
        </article>
      </section>

      <section className="panel leasingRegister">
        <header className="panelHeader"><div><p className="panelKicker">REQUEST REGISTER</p><h2>Maintenance requests</h2></div><span className="countBadge">{requests.length}</span></header>
        {requests.length === 0 ? <div className="emptyState"><h3>No maintenance requests</h3><p>New issues will appear here with priority and lifecycle state.</p></div> : (
          <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Issue</th><th>Location</th><th>Category</th><th>Priority</th><th>Status</th><th>Control</th></tr></thead><tbody>
            {requests.map((item) => (
              <tr key={item.id}>
                <td><strong>{item.title}</strong><span className="recordMeta">{item.description}</span></td>
                <td>{item.propertyName}<span className="recordMeta">{item.unitLabel || "Property-level"}{item.tenantName ? ` · ${item.tenantName}` : ""}</span></td>
                <td>{item.category}</td><td>{item.priority}</td><td><span className={`statusPill status-${item.status}`}>{item.status}</span></td>
                <td>{!["resolved", "cancelled"].includes(item.status) ? (
                  <form action={updateRequestStatusAction} className="stackedForm"><input type="hidden" name="requestId" value={item.id} /><select name="status" defaultValue={item.status === "open" ? "triaged" : item.status === "triaged" ? "in_progress" : "cancelled"}>{item.status === "open" ? <option value="triaged">Mark triaged</option> : null}{item.status === "triaged" ? <option value="in_progress">Mark in progress</option> : null}<option value="cancelled">Cancel</option></select><button className="primaryButton" type="submit">Update</button></form>
                ) : "—"}</td>
              </tr>
            ))}
          </tbody></table></div>
        )}
      </section>

      <section className="panel leasingRegister">
        <header className="panelHeader"><div><p className="panelKicker">WORK ORDER REGISTER</p><h2>Execution</h2></div><span className="countBadge">{workOrders.length}</span></header>
        {workOrders.length === 0 ? <div className="emptyState"><h3>No work orders</h3><p>Dispatch a request to create the operational job record.</p></div> : (
          <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Work</th><th>Location</th><th>Vendor</th><th>Evidence</th><th>Status</th><th>Control / proof</th></tr></thead><tbody>
            {workOrders.map((item) => (
              <tr key={item.id}>
                <td><strong>{item.summary}</strong><span className="recordMeta">{item.requestTitle}</span></td>
                <td>{item.propertyName}<span className="recordMeta">{item.unitLabel || "Property-level"}</span></td>
                <td>{item.vendorName || "Internal / unassigned"}</td><td>{evidenceByWorkOrder.get(item.id) ?? 0}</td><td><span className={`statusPill status-${item.status}`}>{item.status}</span></td>
                <td>
                  {!["completed", "cancelled"].includes(item.status) ? <>
                    <form action={createEvidenceAction} className="stackedForm"><input type="hidden" name="workOrderId" value={item.id} /><input name="note" required placeholder="Completion / inspection note" /><button className="primaryButton" type="submit">Add evidence</button></form>
                    <form action={updateWorkOrderStatusAction} className="stackedForm"><input type="hidden" name="workOrderId" value={item.id} /><select name="status" defaultValue={item.status === "in_progress" ? "completed" : "in_progress"}><option value="in_progress">Start work</option><option value="completed">Complete</option><option value="cancelled">Cancel</option></select><button className="primaryButton" type="submit">Update job</button></form>
                  </> : "—"}
                </td>
              </tr>
            ))}
          </tbody></table></div>
        )}
      </section>

      <section className="panel leasingRegister">
        <header className="panelHeader"><div><p className="panelKicker">QUOTE REGISTER</p><h2>Vendor approvals</h2></div><span className="countBadge">{quotes.length}</span></header>
        {quotes.length === 0 ? <div className="emptyState"><h3>No quotes</h3><p>Submitted vendor costs will stay pending until an authorized approver decides them.</p></div> : (
          <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Scope</th><th>Vendor</th><th>Amount</th><th>Status</th><th>Decision</th></tr></thead><tbody>
            {quotes.map((item) => (
              <tr key={item.id}>
                <td><strong>{item.workOrderSummary}</strong><span className="recordMeta">{item.scopeSummary}</span></td><td>{item.vendorName}</td><td>{formatMoney(item.amountMinor, item.currency)}</td><td><span className={`statusPill status-${item.status}`}>{item.status}</span></td>
                <td>{item.status === "submitted" ? <form action={decideQuoteAction} className="stackedForm"><input type="hidden" name="quoteId" value={item.id} /><select name="decision" defaultValue="approve"><option value="approve">Approve</option><option value="reject">Reject</option></select><button className="primaryButton" type="submit">Record decision</button></form> : item.reviewedAt ? new Date(item.reviewedAt).toLocaleString("en-LK") : "—"}</td>
              </tr>
            ))}
          </tbody></table></div>
        )}
      </section>
    </AppShell>
  );
}
