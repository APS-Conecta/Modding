// SPDX-License-Identifier: AGPL-3.0-or-later
package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/APS-Conecta/Modding/internal/config"
	_ "modernc.org/sqlite"
)

var ErrConflict = errors.New("configuration changed in another request")

type Store struct{ db *sql.DB }
type Snapshot struct {
	Config   config.Config
	Revision int64
}

func Open(directory string) (*Store, error) {
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("state directory must be a private directory (0700)")
	}
	path, err := filepath.Abs(filepath.Join(directory, "modding.db"))
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err == nil {
		err = file.Close()
	} else if errors.Is(err, os.ErrExist) {
		var existing os.FileInfo
		existing, err = os.Lstat(path)
		if err == nil && (!existing.Mode().IsRegular() || existing.Mode().Perm()&0077 != 0) {
			err = fmt.Errorf("state database must be a private regular file (0600)")
		}
	}
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	fail := func(err error) (*Store, error) { db.Close(); return nil, err }
	if _, err = db.Exec("PRAGMA busy_timeout = 3000"); err != nil {
		return fail(err)
	}
	var version int
	if err = db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fail(err)
	}
	if version > 1 {
		return fail(fmt.Errorf("database version %d requires a newer Modding build", version))
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS configuration (id INTEGER PRIMARY KEY CHECK (id=1), revision INTEGER NOT NULL, payload TEXT NOT NULL);`); err != nil {
		return fail(err)
	}
	empty, _ := json.Marshal(config.Config{SchemaVersion: 1, Status: "draft"})
	if _, err = db.Exec("INSERT OR IGNORE INTO configuration VALUES (1,0,?)", string(empty)); err != nil {
		return fail(err)
	}
	if _, err = db.Exec("PRAGMA user_version = 1"); err != nil {
		return fail(err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Load(ctx context.Context) (Snapshot, error) {
	var snapshot Snapshot
	var payload string
	err := s.db.QueryRowContext(ctx, "SELECT revision,payload FROM configuration WHERE id=1").Scan(&snapshot.Revision, &payload)
	if err != nil {
		return snapshot, err
	}
	err = json.Unmarshal([]byte(payload), &snapshot.Config)
	if err == nil && snapshot.Config.SchemaVersion != 1 {
		err = fmt.Errorf("unsupported configuration schema")
	}
	return snapshot, err
}

func (s *Store) Save(ctx context.Context, expected int64, value config.Config) (int64, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return 0, err
	}
	var revision int64
	err = s.db.QueryRowContext(ctx, "UPDATE configuration SET revision=revision+1,payload=? WHERE id=1 AND revision=? RETURNING revision", string(data), expected).Scan(&revision)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrConflict
	}
	return revision, err
}
