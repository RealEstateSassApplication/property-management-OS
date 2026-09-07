import { AppShell } from "../components/app-shell";
import { listTenancies, type Tenancy } from "../lib/property-os";
import {
  InspectionConfigurationError,
  listInspectionItems,
  listInspections,
  type Inspection,
  type InspectionItem,
} from "../lib/inspections";
import {
  acknowledgeInspectionAction,
  completeInspectionAction,
  createInspectionAction,
  createInspectionItemAction,
} from "./actions";

export const dynamic = "force-dynamic";

export default async function InspectionsPage() {
  let inspections: Inspection[] = [];
  let tenancies: Tenancy[] = [];
  const itemsByInspection = new Map<string, InspectionItem[]>();
  let errorMessage = "";

  try {
    [inspections, tenancies] = await Promise.all([listInspections(), listTenancies()]);
    await Promise.all(inspections.map(async (inspection) => {
      itemsByInspection.set(inspection.id, await listInspectionItems(inspection.id));
    }));
  } catch (error) {
    errorMessage = error instanceof InspectionConfigurationError
      ? error.message
      : "Inspection data could not be loaded. Confirm migration 000012 and the Go API are running.";
  }

  const open = inspections.filter((item) => item.status === "draft" || item.status === "in_progress").length;
  const completed = inspections.filter((item) => item.status === "completed").length;
  const acknowledged = inspections.filter((item) => item.status === "acknowledged").length;
  const activeTenancies = tenancies.filter((item) => item.status === "active" || item.status === "upcoming");

  return (
    <AppShell section="Inspections">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">CONDITION CONTROL</p>
          <h1 className="pageTitle">Inspections</h1>
          <p className="pageIntro">Run move-in, move-out and periodic condition reviews with structured findings, document evidence, completion control and acknowledgement.</p>
        </div>
        <div className="metricStrip">
          <div><strong>{open}</strong><span>Open</span></div>
          <div><strong>{completed}</strong><span>Awaiting acknowledgement</span></div>
          <div><strong>{acknowledged}</strong><span>Acknowledged</span></div>
        </div>
      </header>

      {errorMessage ? <div className="noticePanel"><strong>Inspections unavailable</strong><p>{errorMessage}</p></div> : null}

      <div className="leasingGrid">
        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">SCHEDULE</p><h2>Create inspection</h2></div></div>
          <form className="stackedForm" action={createInspectionAction}>
            <label>Tenancy<select name="tenancyId" required defaultValue=""><option value="" disabled>Select tenancy</option>{activeTenancies.map((item) => <option key={item.id} value={item.id}>{item.primaryTenantName} · {item.propertyName} · {item.unitLabel}</option>)}</select></label>
            <div className="formPair"><label>Type<select name="inspectionType" defaultValue="periodic"><option value="move_in">Move-in</option><option value="move_out">Move-out</option><option value="periodic">Periodic</option></select></label><label>Scheduled for<input name="scheduledFor" type="datetime-local" /></label></div>
            <label>Summary<input name="summary" placeholder="Quarterly condition review" /></label>
            <button className="primaryButton" disabled={!activeTenancies.length}>Create inspection</button>
          </form>
        </section>

        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">FINDING</p><h2>Add condition item</h2></div></div>
          <form className="stackedForm" action={createInspectionItemAction}>
            <label>Inspection<select name="inspectionId" required defaultValue=""><option value="" disabled>Select open inspection</option>{inspections.filter((item) => item.status === "draft" || item.status === "in_progress").map((item) => <option key={item.id} value={item.id}>{item.primaryTenantName} · {item.propertyName} · {item.inspectionType.replaceAll("_", " ")}</option>)}</select></label>
            <div className="formPair"><label>Area<input name="area" required placeholder="Kitchen" /></label><label>Item<input name="itemName" required placeholder="Sink and plumbing" /></label></div>
            <label>Condition<select name="condition" defaultValue="good"><option value="good">Good</option><option value="fair">Fair</option><option value="poor">Poor</option><option value="damaged">Damaged</option><option value="not_applicable">Not applicable</option></select></label>
            <label>Notes<input name="notes" placeholder="Condition detail" /></label>
            <label>Evidence document ID<input name="evidenceDocumentId" placeholder="Optional available document UUID" /></label>
            <button className="primaryButton" disabled={!open}>Add finding</button>
          </form>
        </section>
      </div>

      <section className="panel leasingRegister">
        <div className="panelHeader"><div><p className="panelKicker">REGISTER</p><h2>Inspection history</h2></div><span className="countBadge">{inspections.length}</span></div>
        {inspections.length ? <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Inspection</th><th>Property</th><th>Status</th><th>Findings</th><th>Summary</th><th>Control</th></tr></thead><tbody>{inspections.map((inspection) => {
          const items = itemsByInspection.get(inspection.id) ?? [];
          return <tr key={inspection.id}><td><strong className="capitalize">{inspection.inspectionType.replaceAll("_", " ")}</strong><span className="recordMeta">{inspection.primaryTenantName}</span></td><td>{inspection.propertyName}<span className="recordMeta">{inspection.unitLabel}</span></td><td><span className={`statusPill status-${inspection.status}`}>{inspection.status.replaceAll("_", " ")}</span></td><td>{items.length}<span className="recordMeta">{items.filter((item) => item.condition === "damaged" || item.condition === "poor").length} attention items</span></td><td>{inspection.summary || "—"}</td><td>{inspection.status === "draft" || inspection.status === "in_progress" ? <form className="inlineForm" action={completeInspectionAction}><input type="hidden" name="inspectionId" value={inspection.id} /><input name="summary" placeholder="Completion summary" /><button className="tableButton">Complete</button></form> : inspection.status === "completed" ? <form action={acknowledgeInspectionAction}><input type="hidden" name="inspectionId" value={inspection.id} /><button className="tableButton">Acknowledge</button></form> : <span className="recordMeta">Closed</span>}</td></tr>;
        })}</tbody></table></div> : <div className="emptyState"><h3>No inspections</h3><p>Create the first tenancy inspection to begin condition history.</p></div>}
      </section>
    </AppShell>
  );
}
