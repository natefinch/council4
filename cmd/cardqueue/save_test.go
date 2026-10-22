package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveManifestReplacesCompleteCheckpointAndRefusesInterveningEdit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.json")
	original := []byte(`{"checkpoint":"original"}`)
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	replacement := []byte("{\"checkpoint\":\"complete\"}\n")
	if err := saveManifest(path, replacement, original); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(replacement) {
		t.Fatal("complete checkpoint not persisted", err)
	}
	if err := saveManifest(path, original, original); err == nil {
		t.Fatal("stale mutation overwrote an intervening edit")
	}
	got, err = os.ReadFile(path)
	if err != nil || string(got) != string(replacement) {
		t.Fatal("failed write changed existing checkpoint", err)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 {
		t.Fatal("temporary checkpoint was not cleaned up", err)
	}
	missing := filepath.Join(filepath.Dir(path), "missing.json")
	if err := saveManifest(missing, replacement, nil); err == nil {
		t.Fatal("ordinary mutation silently recreated a missing checkpoint")
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Fatal("failed mutation left a new manifest", err)
	}
}
