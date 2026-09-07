import { AppShell } from "../components/app-shell";
import {
  AgentActionConfigurationError,
  listAgentActions,
  type AgentActionRecord,
} from "../lib/agent-actions";
import { decideAgentActionAction } from "./actions";

export const dynamic = "force-dynamic";

function formatMoney(amountMinor: number, currency: string) {
  return new Intl.NumberFormat("en-LK", { style: "currency", currency, maximumFractionDigits: 2 }).format(amountMinor / 100);
}

export default async function AgentActionsPage() {
  let actions: AgentActionRecord[] = [];
  let errorMessage = "";

  try {
    actions = await listAgentActions();
  } catch (error) {
    errorMessage =
      error instanceof AgentActionConfigurationError
        ? error.message
        : "Agent action requests could not be loaded. Confirm migration 000008 and the Go API are running.";
  }

  const proposed = actions.filter((item) => item.status === "proposed");
  const executed = actions.filter((item) => item.status === "executed").length;
  const rejected = actions.filter((item) => item.status === "rejected").length;
  const failed = actions.filter((item) => item.status === "failed").length;

  return (
    <AppShell section="Agent Review">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">HUMAN-IN-THE-LOOP CONTROL</p>
          <h1 className="pageTitle">Agent Review</h1>
          <p className="pageIntro">
            Agents may propose high-risk actions with evidence and reasoning, but financial approvals are executed only after an authorized human decision.
          </p>
        </div>
        <div className="metricStrip" aria-label="Agent action metrics">
          <div><strong>{proposed.length}</strong><span>Awaiting review</span></div>
          <div><strong>{executed}</strong><span>Executed</span></div>
          <div><strong>{rejected}</strong><span>Rejected</span></div>
          <div><strong>{failed}</strong><span>Failed</span></div>
        </div>
      </header>

      {errorMessage ? <div className="noticePanel"><strong>Agent review is unavailable</strong><p>{errorMessage}</p></div> : null}

      <section className="panel leasingRegister">
        <header className="panelHeader">
          <div><p className="panelKicker">PENDING DECISIONS</p><h2>Human approval queue</h2></div>
          <span className="countBadge">{proposed.length}</span>
        </header>
        <div className="dataTableWrap">
          <table className="dataTable">
            <thead><tr><th>Proposal</th><th>Quote</th><th>Agent reasoning</th><th>Review</th></tr></thead>
            <tbody>
              {proposed.map((item) => (
                <tr key={item.id}>
                  <td>
                    <strong>{item.title}</strong>
                    <span className="recordMeta">{item.actionType} · {item.riskLevel} risk</span>
                  </td>
                  <td>
                    <strong>{item.payload.vendorName}</strong>
                    <span className="recordMeta">{formatMoney(item.payload.amountMinor, item.payload.currency)}</span>
                    <span className="recordMeta">{item.payload.workOrderSummary}</span>
                  </td>
                  <td>
                    {item.reasoning}
                    <span className="recordMeta">Scope: {item.payload.scopeSummary}</span>
                  </td>
                  <td>
                    <form action={decideAgentActionAction} className="stackedForm">
                      <input type="hidden" name="actionId" value={item.id} />
                      <label>
                        Human review reason
                        <input name="reason" required minLength={3} placeholder="Why are you approving or rejecting this?" />
                      </label>
                      <div className="formPair">
                        <button className="primaryButton" name="decision" value="approve" type="submit">Approve & execute</button>
                        <button className="primaryButton" name="decision" value="reject" type="submit">Reject</button>
                      </div>
                    </form>
                  </td>
                </tr>
              ))}
              {!proposed.length ? <tr><td colSpan={4}>No agent proposals currently require human review.</td></tr> : null}
            </tbody>
          </table>
        </div>
      </section>

      <section className="panel leasingRegister">
        <header className="panelHeader"><div><p className="panelKicker">AUDIT TRAIL</p><h2>Recent agent actions</h2></div></header>
        <div className="dataTableWrap">
          <table className="dataTable">
            <thead><tr><th>Action</th><th>Resource</th><th>Status</th><th>Decision</th><th>Updated</th></tr></thead>
            <tbody>
              {actions.map((item) => (
                <tr key={item.id}>
                  <td><strong>{item.title}</strong><span className="recordMeta">Proposed by {item.proposedByUserId}</span></td>
                  <td>{item.resourceType}<span className="recordMeta">{item.resourceId}</span></td>
                  <td><span className={`statusPill status-${item.status}`}>{item.status}</span>{item.lastError ? <span className="recordMeta">{item.lastError}</span> : null}</td>
                  <td>{item.decisionReason || "—"}<span className="recordMeta">{item.reviewedByUserId ? `Reviewed by ${item.reviewedByUserId}` : "Awaiting human review"}</span></td>
                  <td>{new Date(item.updatedAt).toLocaleString("en-LK")}</td>
                </tr>
              ))}
              {!actions.length ? <tr><td colSpan={5}>No agent action history yet.</td></tr> : null}
            </tbody>
          </table>
        </div>
      </section>
    </AppShell>
  );
}
