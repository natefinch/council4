package magefiles

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/natefinch/council4/internal/preservation"
)

const corpusPinPath = "testdata/card-preservation/corpus.json"

func pinnedOracleCards() (string, error) {
	pin, err := preservation.Read[preservation.Pin](corpusPinPath)
	if err != nil {
		return "", err
	}
	if err := pin.Validate(); err != nil {
		return "", err
	}
	if path := os.Getenv("COUNCIL4_ORACLE_CARDS"); path != "" {
		return path, pin.Verify(path)
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("locating pinned corpus cache: %w", err)
	}
	path := filepath.Join(cache, "council4", "pinned", pin.JSONSHA256+".json")
	if err := pin.Verify(path); err != nil {
		return "", fmt.Errorf("reading existing pinned corpus (set COUNCIL4_ORACLE_CARDS to a local copy): %w", err)
	}
	return path, nil
}
