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

## Tools

### Read-only

- `portfolio_snapshot` — properties, arrears, open maintenance, near-term lease expiries and queued notifications
- `list_overdue_rent` — overdue obligations and balances
- `list_expiring_leases` — active leases ending within a configurable window
- `list_open_maintenance` — unresolved maintenance requests with optional priority filter

### Mutating

- `queue_rent_reminder` — queues a server-authored, idempotent reminder for a real outstanding obligation
- `create_maintenance_request` — creates a request through the normal API; occupancy and RBAC checks still apply

Mutating tools are deliberately narrow. Generic SQL, arbitrary document access, arbitrary notification sending, quote approval and payment posting are not exposed as MCP tools in this tranche.

## Configuration

Build the server:

```bash
cd apps/api
go build -o property-os-mcp ./cmd/mcp
```

Development environment:

```bash
PROPERTY_OS_API_BASE_URL=http://localhost:8080
PROPERTY_OS_ORGANIZATION_ID=11111111-1111-1111-1111-111111111111
PROPERTY_OS_USER_ID=22222222-2222-2222-2222-222222222222
```

Production should replace `PROPERTY_OS_USER_ID` with a short-lived `PROPERTY_OS_ACCESS_TOKEN` issued for the intended user/service identity. The API remains responsible for authentication and authorization.

A generic MCP client configuration can launch the binary and pass those variables through its `env` configuration. Never commit production access tokens into the repository or an MCP configuration file.
