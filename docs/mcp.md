# Property OS MCP server

Property OS ships a stdio Model Context Protocol server in `apps/api/cmd/mcp`. It uses the official `modelcontextprotocol/go-sdk` and treats the existing authenticated Property OS HTTP API as the sole action boundary.

```text
AI agent / IDE
      |
      v
Property OS MCP (stdio)
      |
      v
HTTP API + Bearer token / development identity
      |
      v
OIDC identity -> organization membership -> RBAC
      |
      v
Property OS domain services -> PostgreSQL
```

The MCP process does not receive a database connection and must not import repositories as an alternative execution path.

## Least-privilege agent identity

Production MCP deployments should use a dedicated Property OS service identity with the `agent` organization role. That role can read the operational context required by the MCP tools and use only the constrained write paths required for reminders, maintenance intake and agent proposals. It cannot manage rent, approve maintenance costs, decide agent proposals, manage documents, create arbitrary notifications or operate the broader maintenance workflow.

The deterministic development seed creates:

```text
User: 17171717-1717-1717-1717-171717171717
Role: agent
```

## Tools

### Read-only

- `portfolio_snapshot` — properties, arrears, open maintenance, near-term lease expiries, queued notifications and pending agent proposals
- `list_overdue_rent` — overdue obligations and balances
- `list_expiring_leases` — active leases ending within a configurable window
- `list_open_maintenance` — unresolved maintenance requests with optional priority filter
- `list_maintenance_quotes` — maintenance quotes for comparison and proposal preparation
- `list_agent_action_requests` — proposal and human-review state; never approves anything

### Constrained mutations

- `queue_rent_reminder` — queues a server-authored, idempotent reminder for a real outstanding obligation
- `create_maintenance_request` — creates a request through the normal API; occupancy and RBAC checks still apply
- `propose_maintenance_quote_approval` — creates a durable high-risk proposal for a submitted quote and snapshots the quote context and agent reasoning

There is intentionally **no MCP quote-approval tool**. A manager, accountant or admin must review the proposal in `/agent-actions`. Approval then invokes the existing maintenance domain service using the human reviewer's identity.

```text
MCP agent
   |
   | propose_maintenance_quote_approval
   v
agent_action_requests (proposed)
   |
   v
Human review workspace
   | approve / reject + reason
   v
maintenance.Service.DecideQuote
   |
   v
executed / rejected / failed audit state
```

Generic SQL, arbitrary document access, arbitrary notification sending, direct quote approval and payment posting are not exposed as MCP tools.

## Configuration

Build the server:

```bash
cd apps/api
go build -o property-os-mcp ./cmd/mcp
```

Development environment using the seeded agent identity:

```bash
PROPERTY_OS_API_BASE_URL=http://localhost:8080
PROPERTY_OS_ORGANIZATION_ID=11111111-1111-1111-1111-111111111111
PROPERTY_OS_USER_ID=17171717-1717-1717-1717-171717171717
```

Production should replace `PROPERTY_OS_USER_ID` with a short-lived `PROPERTY_OS_ACCESS_TOKEN` issued for the dedicated agent/service identity. The API remains responsible for authentication and authorization.

A generic MCP client configuration can launch the binary and pass those variables through its `env` configuration. Never commit production access tokens into the repository or an MCP configuration file.
