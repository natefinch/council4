// Package preservation verifies the integrity of pinned Oracle input data.
package preservation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// Pin identifies immutable input bytes and their provenance.
type Pin struct {
	DownloadURI      string `json:"download_uri"`
	DownloadSHA256   string `json:"download_sha256"`
	JSONSHA256       string `json:"json_sha256"`
	UpdatedAt        string `json:"updated_at"`
	RulesEdition     string `json:"rules_edition"`
	BaselineRevision string `json:"baseline_revision"`
}

// Read decodes one JSON document and refuses trailing documents.
func Read[T any](path string) (T, error) {
	var value T
	file, err := os.Open(path)
	if err != nil {
		return value, fmt.Errorf("opening %s: %w", path, err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decoding %s: %w", path, err)
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return value, fmt.Errorf("%s contains trailing JSON data", path)
	}
	return value, nil
}

// Digest hashes a reader without retaining the corpus in memory.
func Digest(reader io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("hashing input: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Verify checks the actual corpus bytes, not a claimed digest in a report.
func (p Pin) Verify(path string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening pinned corpus: %w", err)
	}
	defer file.Close()
	digest, err := Digest(file)
	if err != nil {
		return err
	}
	if digest != p.JSONSHA256 {
		return fmt.Errorf("corpus digest %s does not match pinned %s", digest, p.JSONSHA256)
	}
	return nil
}

// Validate requires a complete, immutable corpus identity.
func (p Pin) Validate() error {
	for _, digest := range []string{p.DownloadSHA256, p.JSONSHA256} {
		decoded, err := hex.DecodeString(digest)
		if err != nil || len(decoded) != sha256.Size {
			return errors.New("corpus pin requires complete SHA-256 digests")
		}
	}
	if p.DownloadURI == "" || p.UpdatedAt == "" || p.RulesEdition == "" || p.BaselineRevision == "" {
		return errors.New("corpus pin lacks download, date, rules edition or baseline revision")
	}
	return nil
}
