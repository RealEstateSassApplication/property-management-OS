import { AppShell } from "../components/app-shell";
import {
  AccountingConfigurationError,
  getOwnerStatement,
  listDepositAccounts,
  listDepositTransactions,
  listExpenses,
  listPaymentReversals,
  listRentAdjustments,
  type DepositAccount,
  type DepositTransaction,
  type OwnerStatement,
  type PaymentReversal,
  type PropertyExpense,
  type RentAdjustment,
} from "../lib/accounting";
import {
  listLeases,
  listMaintenanceVendors,
  listMaintenanceWorkOrders,
  listOwners,
  listProperties,
  listRentObligations,
  listRentPayments,
  type Lease,
  type MaintenanceVendor,
  type MaintenanceWorkOrder,
  type Owner,
  type Property,
  type RentObligation,
  type RentPayment,
} from "../lib/property-os";
import {
  createAdjustmentAction,
  createDepositAccountAction,
  createDepositTransactionAction,
  createExpenseAction,
  reverseExpenseAction,
  reversePaymentAction,
} from "./actions";

export const dynamic = "force-dynamic";

function formatMoney(amountMinor: number, currency: string) {
  return new Intl.NumberFormat("en-LK", { style: "currency", currency, maximumFractionDigits: 2 }).format(amountMinor / 100);
}

export default async function AccountingPage({ searchParams }: { searchParams: Promise<{ ownerId?: string; from?: string; to?: string }> }) {
  let adjustments: RentAdjustment[] = [];
  let reversals: PaymentReversal[] = [];
  let deposits: DepositAccount[] = [];
  let depositTransactions: DepositTransaction[] = [];
  let expenses: PropertyExpense[] = [];
  let leases: Lease[] = [];
  let owners: Owner[] = [];
  let properties: Property[] = [];
  let obligations: RentObligation[] = [];
  let payments: RentPayment[] = [];
  let vendors: MaintenanceVendor[] = [];
  let workOrders: MaintenanceWorkOrder[] = [];
  let statement: OwnerStatement | null = null;
  let errorMessage = "";

  const params = await searchParams;
  const today = new Date().toISOString().slice(0, 10);
  const monthStart = `${today.slice(0, 7)}-01`;

  try {
    [adjustments, reversals, deposits, depositTransactions, expenses, leases, owners, properties, obligations, payments, vendors, workOrders] = await Promise.all([
      listRentAdjustments(),
      listPaymentReversals(),
      listDepositAccounts(),
      listDepositTransactions(),
      listExpenses(),
      listLeases(),
      listOwners(),
      listProperties(),
      listRentObligations(),
      listRentPayments(),
      listMaintenanceVendors(),
      listMaintenanceWorkOrders(),
    ]);
    if (params.ownerId) {
      statement = await getOwnerStatement(params.ownerId, params.from || monthStart, params.to || today);
    }
  } catch (error) {
    errorMessage = error instanceof AccountingConfigurationError
      ? error.message
      : "Accounting data could not be loaded. Confirm migration 000011 and the Go API are running.";
  }

  const activeLeases = leases.filter((lease) => lease.status === "active");
  const leasesWithoutDeposit = activeLeases.filter((lease) => !deposits.some((account) => account.leaseId === lease.id));
  const liveExpenses = expenses.filter((expense) => !expense.reversed);
  const postedPayments = payments.filter((payment) => payment.status === "posted");
  const depositHeldLkr = deposits.filter((item) => item.currency === "LKR").reduce((sum, item) => sum + item.heldAmountMinor, 0);
  const expensesLkr = liveExpenses.filter((item) => item.currency === "LKR").reduce((sum, item) => sum + item.amountMinor, 0);

  return (
    <AppShell section="Accounting">
      <header className="workspaceHeader">
        <div>
          <p className="eyebrow">FINANCE CONTROL</p>
          <h1 className="pageTitle">Accounting</h1>
          <p className="pageIntro">Control adjustments, reversals, deposit liabilities, property expenses and ownership-weighted statements without mutating historical rent or cash records.</p>
        </div>
        <div className="metricStrip">
          <div><strong>{formatMoney(depositHeldLkr, "LKR")}</strong><span>Deposits held</span></div>
          <div><strong>{formatMoney(expensesLkr, "LKR")}</strong><span>Live expenses</span></div>
          <div><strong>{adjustments.length}</strong><span>Adjustments</span></div>
          <div><strong>{reversals.length}</strong><span>Payment reversals</span></div>
        </div>
      </header>

      {errorMessage ? <div className="noticePanel"><strong>Accounting unavailable</strong><p>{errorMessage}</p></div> : null}

      <div className="leasingGrid">
        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">RECEIVABLE CONTROL</p><h2>Rent adjustment</h2></div></div>
          <form className="stackedForm" action={createAdjustmentAction}>
            <label>Obligation<select name="obligationId" required defaultValue=""><option value="" disabled>Select obligation</option>{obligations.filter((item) => item.state !== "void").map((item) => <option key={item.id} value={item.id}>{item.primaryTenantName} · {item.period} · {formatMoney(item.balanceMinor, item.currency)} balance</option>)}</select></label>
            <div className="formPair"><label>Type<select name="adjustmentType" defaultValue="late_fee"><option value="charge">Charge</option><option value="late_fee">Late fee</option><option value="credit">Credit</option><option value="writeoff">Write-off</option></select></label><label>Amount<input name="amount" inputMode="decimal" required placeholder="5000.00" /></label></div>
            <label>Reason<input name="reason" required placeholder="Approved fee or credit reason" /></label>
            <button className="primaryButton" disabled={!obligations.length}>Post adjustment</button>
          </form>
        </section>

        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">CASH CONTROL</p><h2>Reverse payment</h2></div></div>
          <form className="stackedForm" action={reversePaymentAction}>
            <label>Posted payment<select name="paymentId" required defaultValue=""><option value="" disabled>Select payment</option>{postedPayments.map((payment) => <option key={payment.id} value={payment.id}>{payment.tenantName} · {formatMoney(payment.amountMinor, payment.currency)} · {payment.referenceCode || payment.receivedAt}</option>)}</select></label>
            <label>Reason<input name="reason" required placeholder="Duplicate transfer / returned payment" /></label>
            <p className="recordMeta">Allocations stay in history. The reversal creates compensating receivable adjustments and voids the cash record.</p>
            <button className="primaryButton" disabled={!postedPayments.length}>Reverse payment</button>
          </form>
        </section>
      </div>

      <div className="leasingGrid">
        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">LIABILITY</p><h2>Security deposit account</h2></div></div>
          <form className="stackedForm" action={createDepositAccountAction}>
            <label>Active lease<select name="leaseId" required defaultValue=""><option value="" disabled>Select lease</option>{leasesWithoutDeposit.map((lease) => <option key={lease.id} value={lease.id}>{lease.referenceCode} · {lease.primaryTenantName} · required {formatMoney(lease.depositAmountMinor, lease.currency)}</option>)}</select></label>
            <p className="recordMeta">The required amount and currency are copied from the signed lease by the Go backend.</p>
            <button className="primaryButton" disabled={!leasesWithoutDeposit.length}>Open deposit account</button>
          </form>
        </section>

        <section className="panel">
          <div className="panelHeader"><div><p className="panelKicker">DEPOSIT LEDGER</p><h2>Post deposit movement</h2></div></div>
          <form className="stackedForm" action={createDepositTransactionAction}>
            <label>Deposit account<select name="depositAccountId" required defaultValue=""><option value="" disabled>Select account</option>{deposits.map((item) => <option key={item.id} value={item.id}>{item.tenantName} · held {formatMoney(item.heldAmountMinor, item.currency)}</option>)}</select></label>
            <div className="formPair"><label>Movement<select name="transactionType" defaultValue="received"><option value="received">Received</option><option value="deduction">Deduction</option><option value="refund">Refund</option><option value="adjustment_increase">Adjustment increase</option><option value="adjustment_decrease">Adjustment decrease</option></select></label><label>Amount<input name="amount" inputMode="decimal" required placeholder="50000.00" /></label></div>
            <label>Date<input type="date" name="occurredOn" required defaultValue={today} /></label>
            <label>Note<input name="note" required placeholder="Receipt, deduction or refund reason" /></label>
            <button className="primaryButton" disabled={!deposits.length}>Post deposit movement</button>
          </form>
        </section>
      </div>

      <section className="panel leasingRegister">
        <div className="panelHeader"><div><p className="panelKicker">PROPERTY COSTS</p><h2>Record expense</h2></div></div>
        <form className="stackedForm" action={createExpenseAction}>
          <div className="formPair"><label>Property<select name="propertyId" required defaultValue=""><option value="" disabled>Select property</option>{properties.filter((item) => item.status === "active").map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label><label>Category<select name="category" defaultValue="maintenance"><option value="maintenance">Maintenance</option><option value="utility">Utility</option><option value="tax">Tax</option><option value="insurance">Insurance</option><option value="management">Management</option><option value="cleaning">Cleaning</option><option value="security">Security</option><option value="other">Other</option></select></label></div>
          <div className="formPair"><label>Vendor<select name="vendorId" defaultValue=""><option value="">No vendor</option>{vendors.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label><label>Work order<select name="workOrderId" defaultValue=""><option value="">No work order</option>{workOrders.map((item) => <option key={item.id} value={item.id}>{item.requestTitle} · {item.summary}</option>)}</select></label></div>
          <div className="formPair"><label>Amount (LKR)<input name="amount" inputMode="decimal" required placeholder="18500.00" /></label><label>Incurred<input type="date" name="incurredOn" required defaultValue={today} /></label></div>
          <div className="formPair"><label>Reference<input name="referenceCode" placeholder="INV-2026-001" /></label><label>Currency<input name="currency" defaultValue="LKR" maxLength={3} required /></label></div>
          <label>Note<input name="note" required placeholder="What was this cost for?" /></label>
          <button className="primaryButton" disabled={!properties.length}>Record expense</button>
        </form>
      </section>

      <section className="panel leasingRegister">
        <div className="panelHeader"><div><p className="panelKicker">OWNER REPORTING</p><h2>Statement</h2></div></div>
        <form className="stackedForm" method="get">
          <label>Owner<select name="ownerId" required defaultValue={params.ownerId ?? ""}><option value="" disabled>Select owner</option>{owners.filter((item) => item.status === "active").map((item) => <option key={item.id} value={item.id}>{item.legalName}</option>)}</select></label>
          <div className="formPair"><label>From<input type="date" name="from" defaultValue={params.from || monthStart} required /></label><label>To<input type="date" name="to" defaultValue={params.to || today} required /></label></div>
          <button className="primaryButton" disabled={!owners.length}>Generate statement</button>
        </form>
        {statement ? <>
          <div className="metricStrip accountingSummary">{statement.summaries.map((summary) => <div key={summary.currency}><strong>{formatMoney(summary.netOwnerAmountMinor, summary.currency)}</strong><span>{summary.currency} net · income {formatMoney(summary.incomeMinor, summary.currency)} · expenses {formatMoney(summary.expenseMinor, summary.currency)}</span></div>)}</div>
          <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Date</th><th>Type</th><th>Property</th><th>Description</th><th>Gross</th><th>Owner share</th></tr></thead><tbody>{statement.lines.map((line, index) => <tr key={`${line.date}:${line.propertyId}:${index}`}><td>{line.date}</td><td className="capitalize">{line.lineType}</td><td>{line.propertyName}</td><td>{line.description}</td><td>{formatMoney(line.grossMinor, line.currency)}</td><td>{formatMoney(line.ownerMinor, line.currency)}<span className="recordMeta">{(line.ownershipBps / 100).toFixed(2)}%</span></td></tr>)}</tbody></table></div>
        </> : null}
      </section>

      <section className="panel leasingRegister">
        <div className="panelHeader"><div><p className="panelKicker">DEPOSIT REGISTER</p><h2>Security deposits</h2></div><span className="countBadge">{deposits.length}</span></div>
        {deposits.length ? <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Lease</th><th>Tenant</th><th>Property</th><th>Required</th><th>Held</th></tr></thead><tbody>{deposits.map((item) => <tr key={item.id}><td>{item.leaseReference}</td><td>{item.tenantName}</td><td>{item.propertyName}<span className="recordMeta">{item.unitLabel}</span></td><td>{formatMoney(item.requiredAmountMinor, item.currency)}</td><td>{formatMoney(item.heldAmountMinor, item.currency)}</td></tr>)}</tbody></table></div> : <div className="emptyState"><h3>No deposit accounts</h3><p>Open one from an active lease to start liability tracking.</p></div>}
      </section>

      <section className="panel leasingRegister">
        <div className="panelHeader"><div><p className="panelKicker">EXPENSE REGISTER</p><h2>Property expenses</h2></div><span className="countBadge">{expenses.length}</span></div>
        {expenses.length ? <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Date</th><th>Property</th><th>Category</th><th>Vendor</th><th>Amount</th><th>Status</th><th>Control</th></tr></thead><tbody>{expenses.map((item) => <tr key={item.id}><td>{item.incurredOn}</td><td>{item.propertyName}<span className="recordMeta">{item.referenceCode || item.note}</span></td><td className="capitalize">{item.category}</td><td>{item.vendorName || "—"}</td><td>{formatMoney(item.amountMinor, item.currency)}</td><td><span className={`statusPill status-${item.reversed ? "cancelled" : "active"}`}>{item.reversed ? "reversed" : "posted"}</span></td><td>{item.reversed ? <span className="recordMeta">{item.reversalReason}</span> : <form className="inlineForm" action={reverseExpenseAction}><input type="hidden" name="expenseId" value={item.id} /><input name="reason" required placeholder="Reversal reason" /><button className="tableButton tableButtonDanger">Reverse</button></form>}</td></tr>)}</tbody></table></div> : <div className="emptyState"><h3>No expenses</h3><p>Record property costs to include them in owner statements.</p></div>}
      </section>

      <section className="panel leasingRegister">
        <div className="panelHeader"><div><p className="panelKicker">AUDITABLE CHANGES</p><h2>Adjustments & reversals</h2></div><span className="countBadge">{adjustments.length + reversals.length + depositTransactions.length}</span></div>
        <div className="dataTableWrap"><table className="dataTable"><thead><tr><th>Date</th><th>Type</th><th>Subject</th><th>Amount</th><th>Reason / note</th></tr></thead><tbody>
          {adjustments.map((item) => <tr key={item.id}><td>{item.createdAt.slice(0, 10)}</td><td className="capitalize">{item.adjustmentType.replaceAll("_", " ")}</td><td>{item.tenantName}<span className="recordMeta">{item.leaseReference}</span></td><td>{formatMoney(item.amountMinor, item.currency)}</td><td>{item.reason}</td></tr>)}
          {reversals.map((item) => <tr key={item.id}><td>{item.reversedAt.slice(0, 10)}</td><td>Payment reversal</td><td>{item.tenantName}<span className="recordMeta">{item.referenceCode || item.paymentId}</span></td><td>{formatMoney(item.amountMinor, item.currency)}</td><td>{item.reason}</td></tr>)}
          {depositTransactions.map((item) => <tr key={item.id}><td>{item.occurredOn}</td><td className="capitalize">Deposit {item.transactionType.replaceAll("_", " ")}</td><td>{item.depositAccountId.slice(0, 8)}</td><td>{formatMoney(item.amountMinor, deposits.find((deposit) => deposit.id === item.depositAccountId)?.currency || "LKR")}</td><td>{item.note}</td></tr>)}
        </tbody></table></div>
      </section>
    </AppShell>
  );
}
