# Payment provider integration

Property Management OS keeps external gateway delivery separate from the internal rent ledger. A provider callback is accepted only after signature verification and idempotency checks, then it creates an ordinary posted `payments` row. Reporting, accounting, portals and allocations therefore continue to use one source of truth.

## Generic HMAC webhook

Endpoint:

```text
POST /api/v1/integrations/payments/generic-hmac/webhook
```

Configure `PAYMENT_WEBHOOK_SECRET` with a high-entropy shared secret. The adapter/provider signs the exact raw request body with HMAC-SHA256 and sends:

```text
X-PropertyOS-Signature: sha256=<lowercase hex digest>
```

The request body is JSON:

```json
{
  "eventId": "gateway-event-123",
  "eventType": "payment.paid",
  "organizationId": "11111111-1111-1111-1111-111111111111",
  "tenantId": "66666666-6666-6666-6666-666666666666",
  "amountMinor": 15000000,
  "currency": "LKR",
  "receivedAt": "2026-09-07",
  "referenceCode": "PAY-123"
}
```

`amountMinor` is always an integer minor-unit amount. `receivedAt` uses `YYYY-MM-DD`. The current generic adapter accepts only `payment.paid`.

## Processing guarantees

`payment_provider_events` stores a SHA-256 fingerprint of every accepted event envelope. `(provider,event_id)` is unique. Re-delivery of the identical signed event is idempotent and returns the existing result. Reuse of the same event ID with a different payload is rejected as a conflict.

Before creating cash, the backend validates that the tenant belongs to the supplied organization. The resulting payment uses method `online`, status `posted`, and the provider reference as `reference_code`. Duplicate references are rejected by the existing organization-scoped payment constraint.

Rejected tenant/reference events remain in the provider event ledger with a rejection reason. Successfully processed events link to the created payment and emit a `payment.provider_received` audit event.

## Provider-specific adapters

PayHere, Stripe or other providers should translate their verified provider-native event into the internal `PaidEvent` contract. Do not move provider-specific signature rules into rent/accounting modules and do not let provider callbacks update rent balances directly.

A provider adapter is responsible for:

1. authenticating the provider callback using the provider's documented scheme;
2. validating merchant/account identifiers and paid status;
3. deriving a stable provider event ID and payment reference;
4. converting the provider amount to integer minor units without floating-point arithmetic;
5. passing the normalized paid event into Property OS ingestion;
6. treating retries as idempotent deliveries.

Refunds, disputes and chargebacks should be represented as explicit future provider event types that drive the existing payment-reversal/accounting workflows. They must never delete the original payment.
