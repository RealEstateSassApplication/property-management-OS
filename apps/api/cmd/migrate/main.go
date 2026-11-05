package main

import (
 "context"
 "errors"
 "flag"
 "fmt"
 "log"
 "os"
 "strings"
 "time"

 "github.com/jackc/pgx/v5"
)

func run() error {
 dir := flag.String("dir", "../../migrations", "numbered SQL migration directory")
 baseline := flag.Int("baseline-through", 0, "last fully applied version for a manually reviewed legacy schema")
 confirm := flag.Bool("confirm-reviewed-schema", false, "confirm backup and manual schema review")
 flag.Parse()
 files, err := loadMigrations(*dir)
 if err != nil { return err }
 dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
 if dsn == "" { return errors.New("DATABASE_URL is required") }
 ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
 defer cancel()
 conn, err := pgx.Connect(ctx, dsn)
 if err != nil { return fmt.Errorf("postgres connection failed: %w", err) }
 defer conn.Close(context.Background())
 // Connection-scoped advisory lock serializes concurrent deploy jobs.
 if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock(68357299142)"); err != nil { return err }
 defer func() {
  unlockCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
  defer stop()
  _, _ = conn.Exec(unlockCtx, "SELECT pg_advisory_unlock(68357299142)")
 }()
 if err := initialize(ctx, conn, files, *baseline, *confirm); err != nil { return err }
 return apply(ctx, conn, files)
}

func main() {
 if err := run(); err != nil { log.Fatalf("migration failed: %v", err) }
}
