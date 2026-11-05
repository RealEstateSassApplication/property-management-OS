# Database migrations: safe rollout

Property OS applies migrations through the Go runner in `apps/api/cmd/migrate`. Use `make migrate-all` after configuring `.env`. The runner maintains `public.schema_migrations` with version, file name, SHA-256 checksum and application time.

## Fresh database or previously tracked deployment

1. Back up the database before an upgrade.
2. Run `make migrate-all` from the repository root, **once per deployment**, before starting API/worker replicas.
3. A second run is a no-op. A file changed after it was applied, a missing or reordered version, and unexpected database versions all stop the rollout.
4. Each migration and its ledger entry are committed in one PostgreSQL transaction. A session-level advisory lock serializes concurrent migration processes.
5. **Never edit an applied migration file.** Add a new numbered `*.up.sql` file with a matching `*.down.sql` file.

## Existing installations created before the ledger

If tables already exist but `schema_migrations` does not, the runner **fails closed**. Do not reset the volume, blindly rerun old DDL, or invent a version number.

1. Take and restore-test a backup.
2. Compare the actual schema against the numbered SQL files; identify the last **fully applied, consecutive** migration and check for partially applied statements.
3. With `DATABASE_URL` exported, run from `apps/api`:

   ```sh
   go run ./cmd/migrate -dir ../../migrations -baseline-through=13 -confirm-reviewed-schema
   ```

   Replace 13 with the actual last fully applied version. This explicitly records already-applied migrations and then applies remaining migrations. **Baselining trusts your manual review; it does not verify every constraint, index or row.** Do not use it for an uncertain or partially upgraded database.

4. Verify the final count and schema:

   ```sql
   SELECT version, name, applied_at FROM public.schema_migrations ORDER BY version;
   SELECT to_regclass('public.push_devices');
   ```

No automatic down migrations are provided by the runner. Use a reviewed, backed-up forward-fix migration or an explicitly approved restore procedure rather than running destructive `make migrate-down` targets in shared environments.

Production releases still require staging restore/upgrade rehearsal, finance integrity checks, OIDC access testing, backups, observability and a rollback plan. Passing migration tests alone is not a production sign-off.
