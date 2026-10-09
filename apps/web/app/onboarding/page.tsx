"use client";

import { useActionState } from "react";
import { AppShell } from "../components/app-shell";
import { previewPropertyCSV, type PreflightState } from "./actions";

const initialState: PreflightState = { status: "idle", message: "" };

export default function OnboardingPage() {
  const [state, action, pending] = useActionState(previewPropertyCSV, initialState);
  const report = state.report;

  return (
    <AppShell section="Onboarding">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">MIGRATION STUDIO / FIRST MILE</p>
          <h1 className="pageTitle">Know your data before moving it.</h1>
          <p className="pageIntro">Run a safe, read-only check on property records from an existing system or an Excel export. See errors before importing anything.</p>
        </div>
      </header>

      <section className="contentGrid">
        <article className="panel panelWide">
          <header className="panelHeader">
            <div><p className="panelKicker">STEP 01 · NO DATABASE CHANGES</p><h2>Property migration preflight</h2></div>
          </header>
          <div style={{ padding: 24 }}>
            <form action={action} className="stackedForm">
              <label htmlFor="portfolioCsv">
                CSV export (maximum 512 KiB, 500 properties)
                <input id="portfolioCsv" name="file" type="file" accept=".csv,text/csv" required />
              </label>
              <button className="primaryButton" type="submit" disabled={pending}>{pending ? "Checking file..." : "Review property data"}</button>
            </form>
            {state.message ? (
              <div className="noticePanel" role={state.status === "error" ? "alert" : "status"}>
                <strong>{state.status === "error" ? "Preflight could not run" : "Preflight result"}</strong>
                <p>{state.message}</p>
              </div>
            ) : null}
          </div>
          {report ? (
            <div style={{ padding: "0 24px 24px" }}>
              <div className="metricStrip" aria-label="CSV readiness statistics">
                <div><strong>{report.totalRows}</strong><span>Rows checked</span></div>
                <div><strong>{report.readyRows}</strong><span>Valid rows</span></div>
                <div><strong>{report.invalidRows}</strong><span>Need correction</span></div>
              </div>
              <h3 style={{ marginTop: 28 }}>Sample of validated rows</h3>
              <div className="dataTableWrap">
                <table className="dataTable">
                  <thead><tr><th>Row</th><th>Property</th><th>Reference</th><th>Type</th><th>City</th><th>Country</th></tr></thead>
                  <tbody>
                    {report.previewRows.map((item) => (
                      <tr key={item.row}>
                        <td>{item.row}</td><td>{item.name}</td><td>{item.referenceCode || "—"}</td>
                        <td>{item.propertyType}</td><td>{item.city}</td><td>{item.countryCode}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {report.previewRows.length === 0 ? <p>No valid rows available to preview.</p> : null}
              {report.issues.length ? (
                <>
                  <h3 style={{ marginTop: 28 }}>Corrections needed ({report.issues.length})</h3>
                  <div className="dataTableWrap">
                    <table className="dataTable">
                      <thead><tr><th>Row</th><th>Field</th><th>Issue</th></tr></thead>
                      <tbody>{report.issues.slice(0, 30).map((issue, index) => (
                        <tr key={`${issue.row}-${issue.code}-${index}`}>
                          <td>{issue.row}</td><td>{issue.field || "CSV"}</td><td>{issue.message}</td>
                        </tr>
                      ))}</tbody>
                    </table>
                  </div>
                  {report.issues.length > 30 ? <p>Showing the first 30 issues. Fix these and run preflight again.</p> : null}
                </>
              ) : null}
            </div>
          ) : null}
        </article>

        <aside className="panel">
          <header className="panelHeader"><div><p className="panelKicker">BEFORE YOU START</p><h2>Migration contract</h2></div></header>
          <div style={{ padding: 24, lineHeight: 1.7, fontSize: 14 }}>
            <p>Export your spreadsheet as UTF-8 CSV with headers:</p>
            <code style={{ overflowWrap: "anywhere" }}>referenceCode,name,propertyType,addressLine1,city,region,countryCode,externalAvaraPropertyId</code>
            <p>Required: name, propertyType, addressLine1, city, countryCode.</p>
            <p>Types: house, apartment, building, commercial, land, other.</p>
            <p>Use a two-letter country code like LK. Reference codes and Avara property IDs must be unique within the file.</p>
            <p><strong>Privacy:</strong> The CSV is analyzed for this request and is not saved as an import job. No data is created or overwritten.</p>
            <p>This step validates the file only. Approved imports and reconciliation will be available in a later release.</p>
          </div>
        </aside>
      </section>
    </AppShell>
  );
}
