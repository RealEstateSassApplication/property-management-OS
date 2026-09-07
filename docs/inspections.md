# Inspections

Property Management OS models inspections as tenancy-scoped condition records. Callers select a tenancy; the Go service derives the organization, property and unit relationship server-side so clients cannot attach an inspection to an arbitrary unit.

Supported types are `move_in`, `move_out` and `periodic`. An inspection begins in `draft`, becomes `in_progress` after its first condition item, may be completed only after at least one item exists, and may then be acknowledged. Completed or acknowledged inspections are immutable through the item workflow.

Each condition item records an area, item name, normalized condition (`good`, `fair`, `poor`, `damaged`, `not_applicable`), notes and an optional evidence document. Evidence references must point to an `available` document in the same organization; file bytes remain in S3-compatible object storage under the existing Documents domain.

Manager/admin users can operate the full workflow. Maintenance operators can inspect and acknowledge condition records. Accountants/viewers receive read-only inspection visibility, while owner/tenant portal identities are intentionally not granted organization-wide inspection access; resource-scoped portal inspection access should be exposed through dedicated portal endpoints when needed.

The deterministic development seed includes an acknowledged periodic inspection for the active Maya Silva tenancy with Kitchen and Living room findings.
