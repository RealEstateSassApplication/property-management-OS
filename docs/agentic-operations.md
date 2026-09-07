# Agentic operations and approval policy

Property OS treats AI agents as constrained operators, not administrators. The primary safety boundary is the authenticated HTTP API and organization RBAC; MCP is an adapter over that boundary, not an alternative path into PostgreSQL.

## Risk tiers

| Tier | Examples | Execution policy |
| --- | --- | --- |
| Read-only | portfolio snapshot, arrears, lease expiry, maintenance queue | Agent may execute directly |
| Low-risk write | create maintenance intake, queue server-authored rent reminder | Agent may execute through a purpose-specific API with server-side validation |
| High-risk write | approve maintenance spend, post/reverse financial transactions, terminate leases | Agent may only create a proposal; an authorized human must decide |

The first implemented high-risk workflow is maintenance quote approval.

## Maintenance quote approval

1. The agent reads submitted maintenance quotes through MCP.
2. The agent calls `propose_maintenance_quote_approval` with a quote ID and reasoning.
3. Go reloads the quote from PostgreSQL and rejects quotes that are no longer `submitted`.
4. Property OS snapshots the quote, vendor, amount, currency, work order and scope into `agent_action_requests`.
5. A human reviews the proposal in `/agent-actions` and supplies a mandatory decision reason.
6. Rejecting closes the proposal without changing maintenance state.
7. Approving calls the existing maintenance domain service using the human reviewer's user ID.
8. The proposal is stored as `executed` with the domain result, or `failed` with the execution error.

The snapshot is review context, not authority. The maintenance domain remains authoritative at execution time, so a stale proposal cannot bypass quote/work-order business rules.

## Agent role

The `agent` organization role exists so MCP service identities do not require broad manager credentials. It is intentionally denied direct financial mutation, maintenance cost approval, proposal decisions, document management and arbitrary notification creation.

Future MCP tools should follow the same pattern: add a narrow domain-specific permission, preserve server-authoritative state, and route materially consequential actions through `agent_action_requests` rather than exposing the underlying mutation directly.

## Next guarded actions

Good candidates for the same proposal pattern are:

- rent adjustments, reversals and write-offs
- deposit deductions or refunds
- lease termination/non-renewal
- maintenance approvals above an organization-defined threshold
- vendor onboarding or payout changes
- owner disbursement and reconciliation actions
- destructive document lifecycle actions

Each new action type should define its own payload schema, permission requirement, stale-state validation and execution adapter before it is made available to MCP.
