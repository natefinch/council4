package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
)

func saveManifest(path string, data, expected []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".cardqueue-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(data)
	if writeErr == nil {
		writeErr = file.Sync()
	}
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	live, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !bytes.Equal(live, expected) {
		return errors.New("saved queue changed during mutation: reread it; never overwrite another writer")
	}
	// Same-directory rename avoids a truncated checkpoint. This content check
	// detects intervening edits, but does not replace the single-writer contract.
	return os.Rename(file.Name(), path)
}
