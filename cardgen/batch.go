package cardgen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

const ManifestVersion = 1

const (
	BatchStatusPending    = "pending"
	BatchStatusFetched    = "fetched"
	BatchStatusFetchError = "fetch-error"
)

const (
	BatchFileStatusUnknown  = ""
	BatchFileStatusMissing  = "missing"
	BatchFileStatusExisting = "existing"
)

// CardListItem is one parsed card-list entry before Scryfall canonicalization.
type CardListItem struct {
	InputName string `json:"input_name"`
	Quantity  int    `json:"quantity"`
	Section   string `json:"section,omitempty"`
	Line      int    `json:"line"`
}

// Manifest records resumable batch card-generation workflow state.
type Manifest struct {
	Version int            `json:"version"`
	Cards   []ManifestCard `json:"cards"`
}

// ManifestCard is the manifest row for one unique card in a batch.
type ManifestCard struct {
	InputName     string            `json:"input_name"`
	CanonicalName string            `json:"canonical_name,omitempty"`
	Quantity      int               `json:"quantity"`
	Section       string            `json:"section,omitempty"`
	FirstLine     int               `json:"first_line"`
	Status        string            `json:"status"`
	FileStatus    string            `json:"file_status,omitempty"`
	Layout        string            `json:"layout,omitempty"`
	TypeLine      string            `json:"type_line,omitempty"`
	OracleText    string            `json:"oracle_text,omitempty"`
	Faces         []ManifestFace    `json:"faces,omitempty"`
	FetchError    string            `json:"fetch_error,omitempty"`
	FilePath      string            `json:"file_path,omitempty"`
	Issues        []ValidationIssue `json:"issues,omitempty"`
}

// ManifestFace records oracle text for one face of a multi-face Scryfall card.
type ManifestFace struct {
	Name       string `json:"name"`
	TypeLine   string `json:"type_line,omitempty"`
	OracleText string `json:"oracle_text,omitempty"`
}

// ParseCardList parses a plain-text card list into card entries. It accepts
// lines like "1 Sol Ring" and "Sol Ring", ignores blank/comment lines, strips
// inline comments, and tracks simple section headers such as "COMMANDER:".
func ParseCardList(r io.Reader) ([]CardListItem, error) {
	scanner := bufio.NewScanner(r)
	var items []CardListItem
	section := ""
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		if strings.HasPrefix(raw, "#") {
			continue
		}
		if strings.HasPrefix(raw, "//") {
			header := strings.TrimSpace(strings.TrimPrefix(raw, "//"))
			if header != "" {
				section = header
			}
			continue
		}
		if before, after, ok := strings.Cut(raw, ":"); ok && looksLikeSectionHeader(before) {
			section = normalizeSection(before)
			raw = strings.TrimSpace(after)
			if raw == "" {
				continue
			}
		}
		raw = stripInlineComment(raw)
		if raw == "" {
			continue
		}
		quantity, name := parseQuantity(raw)
		if name == "" {
			return nil, fmt.Errorf("line %d: missing card name", lineNumber)
		}
		items = append(items, CardListItem{
			InputName: name,
			Quantity:  quantity,
			Section:   section,
			Line:      lineNumber,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// NewManifestFromItems creates a unique-card manifest from parsed card-list
// entries, preserving the first line and aggregating quantities per section.
func NewManifestFromItems(items []CardListItem) Manifest {
	manifest := Manifest{Version: ManifestVersion}
	seen := map[string]int{}
	for _, item := range items {
		key := strings.ToLower(item.Section) + "\x00" + strings.ToLower(item.InputName)
		if index, ok := seen[key]; ok {
			manifest.Cards[index].Quantity += item.Quantity
			continue
		}
		seen[key] = len(manifest.Cards)
		manifest.Cards = append(manifest.Cards, ManifestCard{
			InputName: item.InputName,
			Quantity:  item.Quantity,
			Section:   item.Section,
			FirstLine: item.Line,
			Status:    BatchStatusPending,
		})
	}
	return manifest
}

// FetchManifest fills manifest card rows with Scryfall data, using cacheDir when
// non-empty. Fetch errors are recorded per row so a batch can continue.
func FetchManifest(manifest *Manifest, cacheDir string) {
	for i := range manifest.Cards {
		card := &manifest.Cards[i]
		if card.Status == BatchStatusFetched && card.CanonicalName != "" {
			continue
		}
		fetched, err := FetchCardCached(card.InputName, cacheDir)
		if err != nil {
			card.Status = BatchStatusFetchError
			card.FetchError = err.Error()
			continue
		}
		card.CanonicalName = fetched.Name
		card.Layout = fetched.Layout
		card.TypeLine = fetched.TypeLine
		card.OracleText = fetched.OracleText
		card.Faces = manifestFaces(fetched)
		card.FetchError = ""
		card.Status = BatchStatusFetched
	}
}

// FetchCardCached fetches one card by exact name and stores/loads Scryfall JSON
// in cacheDir when cacheDir is non-empty.
func FetchCardCached(name string, cacheDir string) (*ScryfallCard, error) {
	if cacheDir == "" {
		return FetchCard(name)
	}
	path := filepath.Join(cacheDir, url.QueryEscape(strings.ToLower(name))+".json")
	if file, err := os.Open(path); err == nil {
		defer file.Close()
		var card ScryfallCard
		if err := json.NewDecoder(file).Decode(&card); err != nil {
			return nil, fmt.Errorf("decoding cached Scryfall card %s: %w", path, err)
		}
		return &card, nil
	}
	card, err := FetchCard(name)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating cached Scryfall card %s: %w", path, err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(card); err != nil {
		return nil, fmt.Errorf("writing cached Scryfall card %s: %w", path, err)
	}
	return card, nil
}

// LoadManifest reads a manifest from path.
func LoadManifest(path string) (Manifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return Manifest{}, err
	}
	defer file.Close()
	var manifest Manifest
	if err := json.NewDecoder(file).Decode(&manifest); err != nil {
		return Manifest{}, err
	}
	if manifest.Version != ManifestVersion {
		return Manifest{}, fmt.Errorf("manifest version %d is not supported by cardgen version %d", manifest.Version, ManifestVersion)
	}
	return manifest, nil
}

// SaveManifest writes manifest to path with stable indentation.
func SaveManifest(path string, manifest Manifest) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmpPath := path + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// MarkExistingFiles records the expected generated card file and marks rows as
// existing or missing relative to repoRoot.
func MarkExistingFiles(manifest *Manifest, repoRoot string) {
	for i := range manifest.Cards {
		card := &manifest.Cards[i]
		name := card.CanonicalName
		if name == "" {
			name = card.InputName
		}
		relPath := ExpectedCardFile(name)
		card.FilePath = relPath
		if _, err := os.Stat(filepath.Join(repoRoot, relPath)); err == nil {
			card.FileStatus = BatchFileStatusExisting
		} else {
			card.FileStatus = BatchFileStatusMissing
		}
	}
}

// ExpectedCardFile returns the generated card source path for a card name.
func ExpectedCardFile(name string) string {
	return filepath.Join("mtg", "cards", CardNameToPackageLetter(name), CardNameToFileName(name)+".go")
}

func manifestFaces(card *ScryfallCard) []ManifestFace {
	if len(card.CardFaces) == 0 {
		return nil
	}
	faces := make([]ManifestFace, 0, len(card.CardFaces))
	for _, face := range card.CardFaces {
		faces = append(faces, ManifestFace{
			Name:       face.Name,
			TypeLine:   face.TypeLine,
			OracleText: face.OracleText,
		})
	}
	return faces
}

func parseQuantity(line string) (int, string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return 0, ""
	}
	quantityText := fields[0]
	if strings.HasSuffix(strings.ToLower(quantityText), "x") {
		quantityText = quantityText[:len(quantityText)-1]
	}
	quantity, err := strconv.Atoi(quantityText)
	if err != nil {
		return 1, strings.TrimSpace(line)
	}
	if quantity <= 0 {
		quantity = 1
	}
	return quantity, strings.TrimSpace(strings.TrimPrefix(line, fields[0]))
}

func stripInlineComment(line string) string {
	for _, marker := range []string{" #", "\t#"} {
		if before, _, ok := strings.Cut(line, marker); ok {
			line = before
		}
	}
	for _, marker := range []string{" //", "\t//"} {
		if before, after, ok := strings.Cut(line, marker); ok {
			after = strings.TrimSpace(after)
			if after == "" || startsLower(after) {
				line = before
			}
		}
	}
	return strings.TrimSpace(line)
}

func startsLower(value string) bool {
	for _, r := range value {
		return unicode.IsLower(r)
	}
	return false
}

func looksLikeSectionHeader(value string) bool {
	value = normalizeSection(value)
	switch value {
	case "Commander", "Main", "Deck", "Sideboard":
		return true
	default:
		return false
	}
}

func normalizeSection(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + strings.ToLower(value[1:])
}
