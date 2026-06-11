package magefiles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const scryfallOracleCardsMetadataURL = "https://api.scryfall.com/bulk-data/oracle-cards"
const scryfallUserAgent = "council4/1.0 (https://github.com/natefinch/council4)"

// CardSupport regenerates card definitions and the repository's card-support documentation.
func CardSupport(ctx context.Context) error {
	corpusPath, err := oracleCardsCachePath()
	if err != nil {
		return err
	}
	if err := ensureOracleCards(ctx, http.DefaultClient, scryfallOracleCardsMetadataURL, corpusPath); err != nil {
		return err
	}
	generatedRoot := filepath.FromSlash(".cardwork/card-support-generated")
	if err := os.RemoveAll(generatedRoot); err != nil {
		return fmt.Errorf("removing previous card-support generated tree: %w", err)
	}
	if err := os.MkdirAll(".cardwork", 0o750); err != nil {
		return fmt.Errorf("creating cardgen work directory: %w", err)
	}
	return runCommand(
		ctx,
		"go", "run", "./cardgen/oracle/cmd/compilecards",
		"-in", corpusPath,
		"-out", generatedRoot,
		"-report", filepath.FromSlash(".cardwork/current-report.json"),
		"-supported", "supported.md",
		"-unsupported", "unsupported.md",
		"-readme", "README.md",
	)
}

func oracleCardsCachePath() (string, error) {
	if path := os.Getenv("COUNCIL4_ORACLE_CARDS"); path != "" {
		return path, nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("locating user cache directory: %w", err)
	}
	return filepath.Join(cache, "council4", "oracle-cards.json"), nil
}

type bulkDataMetadata struct {
	DownloadURI string `json:"download_uri"`
}

func ensureOracleCards(ctx context.Context, client *http.Client, metadataURL, path string) error {
	info, err := os.Stat(path)
	switch {
	case err == nil && info.Mode().IsRegular() && info.Size() > 0:
		_, _ = fmt.Fprintf(os.Stdout, "Using cached Scryfall Oracle Cards corpus: %s\n", path)
		return nil
	case err == nil:
		return fmt.Errorf("scryfall Oracle Cards cache %s is not a non-empty regular file", path)
	case !errors.Is(err, os.ErrNotExist):
		return fmt.Errorf("checking Scryfall Oracle Cards cache: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating Scryfall bulk-data request: %w", err)
	}
	setScryfallHeaders(request)
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("requesting Scryfall bulk-data metadata: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("requesting Scryfall bulk-data metadata: %s", response.Status)
	}
	var metadata bulkDataMetadata
	if err := json.NewDecoder(response.Body).Decode(&metadata); err != nil {
		return fmt.Errorf("decoding Scryfall bulk-data metadata: %w", err)
	}
	if metadata.DownloadURI == "" {
		return errors.New("scryfall bulk-data metadata has no download URI")
	}

	request, err = http.NewRequestWithContext(ctx, http.MethodGet, metadata.DownloadURI, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating Scryfall Oracle Cards download request: %w", err)
	}
	setScryfallHeaders(request)
	response, err = client.Do(request)
	if err != nil {
		return fmt.Errorf("downloading Scryfall Oracle Cards corpus: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading Scryfall Oracle Cards corpus: %s", response.Status)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("creating Scryfall cache directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), "oracle-cards-*.json")
	if err != nil {
		return fmt.Errorf("creating temporary Scryfall cache: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	written, err := io.Copy(temporary, response.Body)
	if err != nil {
		_ = temporary.Close()
		return fmt.Errorf("writing Scryfall Oracle Cards cache: %w", err)
	}
	if written == 0 {
		_ = temporary.Close()
		return errors.New("scryfall Oracle Cards download was empty")
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("closing Scryfall Oracle Cards cache: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("installing Scryfall Oracle Cards cache: %w", err)
	}
	_, _ = fmt.Fprintf(os.Stdout, "Downloaded Scryfall Oracle Cards corpus: %s\n", path)
	return nil
}

func setScryfallHeaders(request *http.Request) {
	request.Header.Set("User-Agent", scryfallUserAgent)
	request.Header.Set("Accept", "application/json")
}

func runCommand(ctx context.Context, name string, args ...string) error {
	command := exec.CommandContext(ctx, name, args...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("running %s: %w", name, err)
	}
	return nil
}
