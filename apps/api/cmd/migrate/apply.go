package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
)

func apply(ctx context.Context, conn *pgx.Conn, files []migration) error {
	rows, err := conn.Query(ctx, "SELECT version, name, checksum FROM public.schema_migrations ORDER BY version")
	if err != nil {
		return err
	}
	type prior struct{ name, checksum string }
	applied := map[int]prior{}
	for rows.Next() {
		var version int
		var name, checksum string
		if err := rows.Scan(&version, &name, &checksum); err != nil {
			rows.Close()
			return err
		}
		applied[version] = prior{name, strings.TrimSpace(checksum)}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if len(applied) > len(files) {
		return errors.New("database version newer than this release")
	}
	for v := 1; v <= len(applied); v++ {
		if _, ok := applied[v]; !ok {
			return fmt.Errorf("schema migration ledger gap at version %d", v)
		}
	}
	if len(applied) == 0 {
		var hasUserTables bool
		if err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_tables WHERE schemaname='public' AND tablename NOT IN ('schema_migrations'))").Scan(&hasUserTables); err != nil {
			return err
		}
		if hasUserTables {
			return errors.New("untracked existing tables; refusing migration")
		}
	}
	for _, file := range files {
		if old, exists := applied[file.version]; exists {
			if old.name != file.name || old.checksum != file.checksum {
				return fmt.Errorf("checksum/name mismatch for applied migration %s", file.name)
			}
			continue
		}
		if file.version != len(applied)+1 {
			return fmt.Errorf("out-of-order migration %s", file.name)
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return err
		}
		// Simple protocol handles multi-statement SQL. File and ledger commit together.
		if _, err := tx.Exec(ctx, file.sql, pgx.QueryExecModeSimpleProtocol); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("apply %s: %w", file.name, err)
		}
		if _, err := tx.Exec(ctx, recordSQL, file.version, file.name, file.checksum); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", file.name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", file.name, err)
		}
		applied[file.version] = prior{file.name, file.checksum}
		log.Printf("applied %s", file.name)
	}
	log.Printf("schema verified at version %06d", len(files))
	return nil
}
