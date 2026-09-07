import { AppShell } from "../components/app-shell";
import { DocumentActions, DocumentUpload, type ResourceOption } from "./upload-client";
import { DocumentConfigurationError, listDocuments, type StoredDocument } from "../lib/documents";
import {
  listLeases,
  listMaintenanceVendors,
  listMaintenanceWorkOrders,
  listOwners,
  listProperties,
  listTenants,
} from "../lib/property-os";

export const dynamic = "force-dynamic";

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export default async function DocumentsPage() {
  let documents: StoredDocument[] = [];
  let resources: ResourceOption[] = [{ value: "organization:", label: "Organization · General document" }];
  let errorMessage = "";

  try {
    const [documentRows, properties, leases, tenants, owners, workOrders, vendors] = await Promise.all([
      listDocuments(),
      listProperties(),
      listLeases(),
      listTenants(),
      listOwners(),
      listMaintenanceWorkOrders(),
      listMaintenanceVendors(),
    ]);
    documents = documentRows;
    resources = resources.concat(
      properties.map((item) => ({ value: `property:${item.id}`, label: `Property · ${item.name}` })),
      leases.map((item) => ({ value: `lease:${item.id}`, label: `Lease · ${item.referenceCode} · ${item.primaryTenantName}` })),
      tenants.map((item) => ({ value: `tenant:${item.id}`, label: `Tenant · ${item.legalName}` })),
      owners.map((item) => ({ value: `owner:${item.id}`, label: `Owner · ${item.legalName}` })),
      workOrders.map((item) => ({ value: `work_order:${item.id}`, label: `Work order · ${item.summary}` })),
      vendors.map((item) => ({ value: `vendor:${item.id}`, label: `Vendor · ${item.name}` })),
    );
  } catch (error) {
    errorMessage = error instanceof DocumentConfigurationError
      ? error.message
      : "Document data could not be loaded. Confirm migration 000006, the Go API, and development identity are configured.";
  }

  const available = documents.filter((item) => item.status === "available").length;
  const pending = documents.filter((item) => item.status === "pending").length;
  const quarantined = documents.filter((item) => item.status === "quarantined").length;

  return (
    <AppShell section="Documents">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">CONTROLLED RECORDS</p>
          <h1 className="pageTitle">Documents</h1>
          <p className="pageIntro">Attach leases, IDs, inspection reports, invoices, receipts and operational proof without storing file bytes in PostgreSQL.</p>
        </div>
        <div className="metricStrip" aria-label="Document metrics">
          <div><strong>{documents.length}</strong><span>Active records</span></div>
          <div><strong>{available}</strong><span>Verified</span></div>
          <div><strong>{pending}</strong><span>Pending</span></div>
          <div><strong>{quarantined}</strong><span>Quarantined</span></div>
        </div>
      </header>

      {errorMessage ? <div className="noticePanel"><strong>Documents are unavailable</strong><p>{errorMessage}</p></div> : null}

      <section className="leasingGrid">
        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">SECURE UPLOAD</p><h2>Add document</h2></div></header>
          <p className="formHint">Maximum 25 MB. Allowed: PDF, JPEG, PNG, WebP, TXT and CSV. The browser uploads directly to configured S3-compatible storage and Go verifies the object before publishing it.</p>
          <DocumentUpload resources={resources} />
        </article>
        <article className="panel">
          <header className="panelHeader"><div><p className="panelKicker">STORAGE POLICY</p><h2>What the system trusts</h2></div></header>
          <div className="detailList">
            <p><strong>Backend-generated keys</strong><span>Clients never choose object paths.</span></p>
            <p><strong>Post-upload verification</strong><span>Size and MIME must match the upload intent.</span></p>
            <p><strong>Quarantine on mismatch</strong><span>Unexpected objects never become downloadable.</span></p>
            <p><strong>Short download grants</strong><span>Available files use five-minute signed URLs.</span></p>
          </div>
        </article>
      </section>

      <section className="panel">
        <header className="panelHeader"><div><p className="panelKicker">REGISTER</p><h2>Document records</h2></div></header>
        <div className="tableWrap">
          <table className="dataTable">
            <thead><tr><th>File</th><th>Type</th><th>Attached to</th><th>Size</th><th>Status</th><th>Added</th><th /></tr></thead>
            <tbody>
              {documents.map((item) => (
                <tr key={item.id}>
                  <td><strong>{item.fileName}</strong><small>{item.contentType}</small></td>
                  <td>{item.kind}</td>
                  <td><span className="statusPill">{item.resourceType}</span><small>{item.resourceId}</small></td>
                  <td>{formatBytes(item.sizeBytes)}</td>
                  <td><span className={`statusPill status-${item.status}`}>{item.status}</span></td>
                  <td>{new Date(item.createdAt).toLocaleDateString("en-LK")}</td>
                  <td><DocumentActions id={item.id} available={item.status === "available"} /></td>
                </tr>
              ))}
              {!documents.length ? <tr><td colSpan={7}>No document records yet.</td></tr> : null}
            </tbody>
          </table>
        </div>
      </section>
    </AppShell>
  );
}
