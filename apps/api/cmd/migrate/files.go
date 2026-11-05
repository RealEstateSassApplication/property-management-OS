package main

import (
 "crypto/sha256"
 "encoding/hex"
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "regexp"
 "sort"
 "strconv"
 "strings"
)

type migration struct {
 version int
 name, sql, checksum string
}
var filePattern = regexp.MustCompile(`^([0-9]{6})_.+\.up\.sql$`)

func loadMigrations(dir string) ([]migration, error) {
 entries, err := os.ReadDir(dir)
 if err != nil { return nil, err }
 files := []migration{}
 for _, entry := range entries {
  if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") { continue }
  name := entry.Name()
  match := filePattern.FindStringSubmatch(name)
  if match == nil { return nil, fmt.Errorf("invalid migration filename: %s", name) }
  version, _ := strconv.Atoi(match[1])
  if version == 0 { return nil, fmt.Errorf("invalid migration version: %s", name) }
  if _, err := os.Stat(filepath.Join(dir, strings.TrimSuffix(name, ".up.sql")+".down.sql")); err != nil {
   return nil, fmt.Errorf("rollback file missing for %s: %w", name, err)
  }
  raw, err := os.ReadFile(filepath.Join(dir, name))
  if err != nil { return nil, err }
  if strings.TrimSpace(string(raw)) == "" { return nil, fmt.Errorf("empty migration: %s", name) }
  sum := sha256.Sum256(raw)
  files = append(files, migration{version, name, string(raw), hex.EncodeToString(sum[:])})
 }
 sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
 if len(files) == 0 { return nil, errors.New("no migrations found") }
 for i, file := range files {
  if file.version != i+1 { return nil, fmt.Errorf("missing/duplicate migration before %s (expected %d)", file.name, i+1) }
 }
 return files, nil
}
