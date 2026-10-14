package magefiles

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/natefinch/council4/internal/preservation"
)

const corpusPinPath = "testdata/card-preservation/corpus.json"

func pinnedOracleCards(ctx context.Context) (string, error) {
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
	if _, err := os.Stat(path); err == nil {
		return path, pin.Verify(path)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("checking pinned corpus cache: %w", err)
	}
	if err := installPinnedCorpus(ctx, http.DefaultClient, pin, path); err != nil {
		return "", err
	}
	return path, nil
}

func installPinnedCorpus(ctx context.Context, client *http.Client, pin preservation.Pin, path string) error {
	if err := pin.Validate(); err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, pin.DownloadURI, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating pinned corpus request: %w", err)
	}
	setScryfallHeaders(request)
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("downloading pinned corpus: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading pinned corpus: %s", response.Status)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("creating pinned corpus directory: %w", err)
	}
	download, err := os.CreateTemp(filepath.Dir(path), "corpus-download-*")
	if err != nil {
		return fmt.Errorf("creating corpus download: %w", err)
	}
	defer download.Close()
	defer os.Remove(download.Name())
	if _, err := io.Copy(download, response.Body); err != nil {
		return fmt.Errorf("saving pinned corpus download: %w", err)
	}
	if _, err := download.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewinding corpus download: %w", err)
	}
	digest, err := preservation.Digest(download)
	if err != nil {
		return err
	}
	if digest != pin.DownloadSHA256 {
		return fmt.Errorf("download digest %s does not match pinned %s", digest, pin.DownloadSHA256)
	}
	if _, err := download.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewinding verified download: %w", err)
	}
	converted, err := os.CreateTemp(filepath.Dir(path), "corpus-json-*")
	if err != nil {
		return fmt.Errorf("creating corpus JSON: %w", err)
	}
	defer converted.Close()
	defer os.Remove(converted.Name())
	if err := writeOracleCardsJSON(converted, download); err != nil {
		return err
	}
	if err := converted.Close(); err != nil {
		return fmt.Errorf("closing corpus JSON: %w", err)
	}
	if err := pin.Verify(converted.Name()); err != nil {
		return err
	}
	if err := os.Rename(converted.Name(), path); err != nil {
		return fmt.Errorf("installing pinned corpus: %w", err)
	}
	return nil
}
