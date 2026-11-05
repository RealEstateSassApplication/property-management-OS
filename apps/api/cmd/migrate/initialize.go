package main

import (
 "context"
 "errors"
 "fmt"

 "github.com/jackc/pgx/v5"
)

const ledgerDDL = "CREATE TABLE public.schema_migrations (version INTEGER PRIMARY KEY CHECK (version > 0), name TEXT NOT NULL, checksum CHAR(64) NOT NULL, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())"
const recordSQL = "INSERT INTO public.schema_migrations (version, name, checksum) VALUES ($1, $2, $3)"

func initialize(ctx context.Context, conn *pgx.Conn, files []migration, baseline int, confirm bool) error {
 var hasLedger, hasUserTables bool
 err := conn.QueryRow(ctx, "SELECT to_regclass('public.schema_migrations') IS NOT NULL, EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename <> 'schema_migrations')").Scan(&hasLedger, &hasUserTables)
 if err != nil { return err }
 if hasLedger {
  if baseline != 0 || confirm { return errors.New("cannot baseline a database with an existing migration ledger") }
  return nil
 }
 if hasUserTables && (baseline == 0 || !confirm) {
  return errors.New("existing tables without migration ledger: back up and review, then explicitly baseline (docs/migrations.md)")
 }
 if !hasUserTables && (baseline != 0 || confirm) { return errors.New("cannot baseline a fresh database") }
 if baseline < 0 || baseline > len(files) { return fmt.Errorf("invalid baseline version %d", baseline) }
 if hasUserTables {
  var core bool
  if err := conn.QueryRow(ctx, "SELECT to_regclass('public.organizations') IS NOT NULL").Scan(&core); err != nil { return err }
  if !core { return errors.New("cannot baseline: organizations table missing") }
  if baseline == len(files) {
   var push bool
   if err := conn.QueryRow(ctx, "SELECT to_regclass('public.push_devices') IS NOT NULL").Scan(&push); err != nil { return err }
   if !push { return errors.New("cannot baseline latest version: push_devices table missing") }
  }
 }
 tx, err := conn.Begin(ctx)
 if err != nil { return err }
 defer tx.Rollback(ctx)
 if _, err := tx.Exec(ctx, ledgerDDL); err != nil { return err }
 for _, file := range files[:baseline] {
  if _, err := tx.Exec(ctx, recordSQL, file.version, file.name, file.checksum); err != nil { return err }
 }
 return tx.Commit(ctx)
}
