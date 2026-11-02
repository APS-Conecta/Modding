package state

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/APS-Conecta/Modding/internal/catalog"
	"github.com/APS-Conecta/Modding/internal/config"
)

func TestPersistenceAndConcurrentEdits(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "state")
	ctx := context.Background()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := s.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if initial.Config.Establishment.Code != "" || initial.Revision != 0 {
		t.Fatal("preselected establishment")
	}
	value := config.Config{SchemaVersion: 1, Establishment: catalog.Entry{Code: "900001", Name: "Centro de prueba"}, Status: "draft"}
	var wg sync.WaitGroup
	errorsC := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := s.Save(ctx, 0, value); errorsC <- err }()
	}
	wg.Wait()
	close(errorsC)
	wins := 0
	for err := range errorsC {
		if err == nil {
			wins++
		} else if !errors.Is(err, ErrConflict) {
			t.Fatal(err)
		}
	}
	if wins != 1 {
		t.Fatalf("concurrent winners: %d", wins)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s, err = Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	saved, err := s.Load(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Revision != 1 || saved.Config.Establishment.Name != value.Establishment.Name {
		t.Fatalf("restart lost state: %#v", saved)
	}
	info, err := os.Stat(filepath.Join(dir, "modding.db"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("database permissions: %o", info.Mode().Perm())
	}
}

func TestUnsafeStateLocationRefused(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if s, err := Open(dir); err == nil {
		s.Close()
		t.Fatal("accepted shared state directory")
	}
	dir = filepath.Join(t.TempDir(), "state")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "elsewhere"), filepath.Join(dir, "modding.db")); err != nil {
		t.Fatal(err)
	}
	if s, err := Open(dir); err == nil {
		s.Close()
		t.Fatal("accepted database symlink")
	}
}
