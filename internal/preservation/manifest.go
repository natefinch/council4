// Package preservation records independent compiler and registry inventories.
package preservation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
)

// Pin identifies immutable input bytes and the rules/code baseline.
type Pin struct {
	DownloadURI      string `json:"download_uri"`
	DownloadSHA256   string `json:"download_sha256"`
	JSONSHA256       string `json:"json_sha256"`
	UpdatedAt        string `json:"updated_at"`
	RulesEdition     string `json:"rules_edition"`
	BaselineRevision string `json:"baseline_revision"`
}

// Record describes a compiler result or an independently registered definition.
// OracleIDs are candidate identities when a legacy registration has only a name.
type Record struct {
	Key               string   `json:"key"`
	Name              string   `json:"name"`
	Layout            string   `json:"layout"`
	Faces             []string `json:"faces"`
	OracleIDs         []string `json:"oracle_ids,omitempty"`
	OracleTextSHA256  string   `json:"oracle_text_sha256,omitempty"`
	DefinitionSHA256  string   `json:"definition_sha256,omitempty"`
	ImplementationIDs []string `json:"implementation_ids,omitempty"`
	Refusals          []string `json:"refusals,omitempty"`
}

// Manifest keeps acceptance separate from registration and structural loading.
// Fingerprints describe generated source or structural CardDef values, not
// certified gameplay behavior.
type Manifest struct {
	SchemaVersion int      `json:"schema_version"`
	Input         Pin      `json:"input"`
	Compiler      []Record `json:"compiler"`
	Registry      []Record `json:"registry"`
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

// Write writes deterministic JSON after the complete artifact is assembled.
func Write[T any](path string, value T) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
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

// Validate rejects ambiguous identities before any regression comparison.
func (m Manifest) Validate() error {
	if m.SchemaVersion != 1 {
		return fmt.Errorf("unsupported preservation schema %d", m.SchemaVersion)
	}
	if err := m.Input.Validate(); err != nil {
		return err
	}
	for _, records := range [][]Record{m.Compiler, m.Registry} {
		seen := make(map[string]bool)
		for _, record := range records {
			if record.Key == "" || record.Name == "" || seen[record.Key] {
				return fmt.Errorf("missing or duplicate preservation identity %q", record.Key)
			}
			seen[record.Key] = true
			if record.DefinitionSHA256 == "" && len(record.Refusals) == 0 {
				return fmt.Errorf("identity %s has neither a fingerprint nor an explicit refusal", record.Key)
			}
			if record.DefinitionSHA256 != "" {
				digest, err := hex.DecodeString(record.DefinitionSHA256)
				if err != nil || len(digest) != sha256.Size || len(record.Refusals) != 0 {
					return fmt.Errorf("identity %s has an invalid/partially refused fingerprint", record.Key)
				}
			}
		}
	}
	return nil
}

// Sort makes inventories independent of corpus/registration iteration order.
func Sort(records []Record) {
	slices.SortFunc(records, func(a, b Record) int {
		if a.Key < b.Key {
			return -1
		}
		if a.Key > b.Key {
			return 1
		}
		return 0
	})
}
