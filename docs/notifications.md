# Notification delivery architecture

Property OS uses a PostgreSQL transactional outbox instead of sending email/SMS/WhatsApp inside HTTP request handlers.

```text
API / agent action
      |
      v
notification_outbox (durable)
      |
      v
Go worker -- FOR UPDATE SKIP LOCKED
      |
      v
DeliveryProvider
  |             |
 log (dev)   webhook (production adapter)
      |
      v
delivered / retry / dead
```

## Guarantees

- organization-scoped rows
- idempotency keys for duplicate-sensitive workflows
- `pending`, `processing`, `retry`, `delivered`, and `dead` states
- attempt count and capped exponential retry delay
- stale processing locks can be reclaimed after ten minutes
- worker concurrency uses `FOR UPDATE SKIP LOCKED`
- API writes return after the outbox row is committed; external delivery is asynchronous
- production workers cannot use the development log provider

## Rent reminders

`POST /api/v1/notifications/rent-reminders` accepts only the obligation ID, delivery channel, and recipient. The Go service loads the actual outstanding balance, due date, tenant, property, and unit from the ledger and authors the reminder itself. Same-day reminders for the same obligation/channel/recipient are idempotent.

This endpoint is the preferred agentic notification action because an agent cannot invent the authoritative balance or due date.
