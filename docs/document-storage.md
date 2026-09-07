# Document Storage

Property Management OS stores document **metadata and authorization relationships in PostgreSQL** and file bytes in an S3-compatible object store. The application never stores binary file bodies in PostgreSQL.

## Upload lifecycle

```text
manager chooses resource + file
        ↓
POST /api/v1/documents/uploads
        ↓
Go validates resource, MIME, size and filename
        ↓
PostgreSQL document(status=pending)
        ↓
10-minute presigned PUT
        ↓
browser uploads bytes directly to object storage
        ↓
POST /api/v1/documents/{id}/complete
        ↓
Go HEADs object storage
        ↓
size + content type match?
       / \
     yes  no
      ↓    ↓
available quarantined
```

The backend generates every storage key. Clients never choose bucket paths.

## Supported files

Current maximum size: **25 MiB**.

Allowed content types:

- `application/pdf`
- `image/jpeg`
- `image/png`
- `image/webp`
- `text/plain`
- `text/csv`

The MIME value supplied to the upload intent and the MIME returned by object storage must match after normalization. The object size must exactly match the declared size before the record becomes available.

## Resource relationships

Documents can currently be attached to:

- organization
- property
- unit
- tenant
- lease
- owner
- rent payment
- maintenance request
- work order
- vendor

The backend verifies that non-organization resources exist inside the same organization before issuing an upload grant.

## Document kinds

- lease
- identity
- inspection
- invoice
- receipt
- maintenance
- statement
- other

## Authorization

Generalized document routes are intentionally restricted to `admin` and `manager` roles in this tranche. Sensitive files must not automatically inherit the broad read permissions used by operational dashboards.

Owner, tenant, accountant, viewer, and maintenance access should be introduced through **resource-scoped policies** in their respective portal/workflow tranches.

## Download and deletion

Only `available` documents receive download grants. Download URLs expire after five minutes.

Deletion removes the storage object first and then tombstones the PostgreSQL record with `status=deleted`. Lifecycle changes create actor-aware audit events.

## Storage configuration

```bash
STORAGE_BUCKET=property-os-documents
STORAGE_REGION=us-east-1
STORAGE_ENDPOINT=
STORAGE_PATH_STYLE=false
```

The AWS SDK default credential chain is used. This supports AWS S3 and compatible providers such as Cloudflare R2 or Backblaze B2 S3 by setting the provider endpoint. A self-hosted S3-compatible service can also be used by a deployer.

For browser-direct uploads, the bucket must allow CORS from the Property OS web origin for the required `PUT` headers. Keep bucket objects private; presigned grants are the access mechanism.

## Security follow-ups

Before broad production rollout, add malware scanning/quarantine promotion, configurable retention policies, optional checksum verification, object-versioning policy, and resource-scoped portal authorization.
