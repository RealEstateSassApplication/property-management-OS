import { AppShell } from "../components/app-shell";
import { listNotifications, NotificationConfigurationError, type NotificationRecord } from "../lib/notifications";

export const dynamic = "force-dynamic";

export default async function NotificationsPage() {
  let notifications: NotificationRecord[] = [];
  let errorMessage = "";
  try {
    notifications = await listNotifications();
  } catch (error) {
    errorMessage = error instanceof NotificationConfigurationError
      ? error.message
      : "Notification data could not be loaded. Confirm migration 000007, the Go API, and identity configuration.";
  }

  const pending = notifications.filter((item) => item.status === "pending" || item.status === "processing").length;
  const retrying = notifications.filter((item) => item.status === "retry").length;
  const dead = notifications.filter((item) => item.status === "dead").length;
  const delivered = notifications.filter((item) => item.status === "delivered").length;

  return (
    <AppShell section="Notifications">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">DELIVERY OPERATIONS</p>
          <h1 className="pageTitle">Notifications</h1>
          <p className="pageIntro">Track durable email, SMS, WhatsApp and webhook delivery without sending messages inside user-facing API requests.</p>
        </div>
        <div className="metricStrip" aria-label="Notification metrics">
          <div><strong>{pending}</strong><span>Queued</span></div>
          <div><strong>{retrying}</strong><span>Retrying</span></div>
          <div><strong>{delivered}</strong><span>Delivered</span></div>
          <div><strong>{dead}</strong><span>Dead letter</span></div>
        </div>
      </header>

      {errorMessage ? <div className="noticePanel"><strong>Notifications are unavailable</strong><p>{errorMessage}</p></div> : null}

      <section className="panel">
        <header className="panelHeader"><div><p className="panelKicker">OUTBOX</p><h2>Delivery register</h2></div></header>
        <p className="formHint">Mutations create durable outbox rows first. A separate Go worker claims them with PostgreSQL row locks, sends through the configured provider, and records delivery/retry/dead-letter state.</p>
        <div className="tableWrap">
          <table className="dataTable">
            <thead><tr><th>Topic</th><th>Channel</th><th>Recipient</th><th>Status</th><th>Attempts</th><th>Resource</th><th>Created</th></tr></thead>
            <tbody>
              {notifications.map((item) => (
                <tr key={item.id}>
                  <td><strong>{item.subject || item.topic}</strong><small>{item.topic}</small></td>
                  <td>{item.channel}</td>
                  <td>{item.recipient}</td>
                  <td><span className={`statusPill status-${item.status}`}>{item.status}</span>{item.lastError ? <small>{item.lastError}</small> : null}</td>
                  <td>{item.attemptCount}/{item.maxAttempts}</td>
                  <td>{item.resourceType ? <><span>{item.resourceType}</span><small>{item.resourceId}</small></> : "—"}</td>
                  <td>{new Date(item.createdAt).toLocaleString("en-LK")}</td>
                </tr>
              ))}
              {!notifications.length ? <tr><td colSpan={7}>No notifications have been queued yet.</td></tr> : null}
            </tbody>
          </table>
        </div>
      </section>
    </AppShell>
  );
}
