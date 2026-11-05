package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeMigration(t *testing.T, dir, name, sql string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".up.sql"), []byte(sql), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".down.sql"), []byte("SELECT 1;"), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadMigrationsValidatesSequence(t *testing.T) {
	dir := t.TempDir()
	writeMigration(t, dir, "000001_core", "CREATE TABLE x (id int);")
	writeMigration(t, dir, "000002_second", "CREATE TABLE y (id int);")
	files, err := loadMigrations(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].version != 1 || files[1].version != 2 {
		t.Fatalf("unexpected files: %#v", files)
	}
	if len(files[0].checksum) != 64 {
		t.Fatal("sha256 checksum missing")
	}
	writeMigration(t, dir, "000004_fourth", "SELECT 1;")
	if _, err := loadMigrations(dir); err == nil {
		t.Fatal("expected gap rejection")
	}
}

func TestLoadMigrationsDetectsTampering(t *testing.T) {
	dir := t.TempDir()
	writeMigration(t, dir, "000001_core", "SELECT 1;")
	first, err := loadMigrations(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000001_core.up.sql"), []byte("SELECT 2;"), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := loadMigrations(dir)
	if err != nil {
		t.Fatal(err)
	}
	if first[0].checksum == second[0].checksum {
		t.Fatal("content tampering must change checksum")
	}
}

func TestLoadMigrationsRejectsMissingDownAndInvalidName(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "000001_core.up.sql"), []byte("SELECT 1;"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadMigrations(dir); err == nil || !strings.Contains(err.Error(), "rollback") {
		t.Fatalf("expected missing down file error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000001_core.down.sql"), []byte("SELECT 1;"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.up.sql"), []byte("SELECT 1;"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadMigrations(dir); err == nil {
		t.Fatal("expected invalid migration name error")
	}
}
