// Package supportinventory joins production registrations with frozen compiler
// evidence. Compiler acceptance never establishes installed behavior.
package supportinventory

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/cards"
)

// OracleFace preserves a printed face name without inventing installed identity.
type OracleFace struct {
	Name string `json:"name"`
}

// OracleCard is the identity subset used from the pinned Oracle corpus.
type OracleCard struct {
	OracleID string       `json:"oracle_id"`
	Name     string       `json:"name"`
	Layout   string       `json:"layout,omitempty"`
	Faces    []OracleFace `json:"card_faces"`
}

// DiagnosticCard retains an exact compiler identity and its maintenance evidence.
type DiagnosticCard struct {
	OracleID    string          `json:"oracle_id"`
	Name        string          `json:"name"`
	Diagnostics json.RawMessage `json:"diagnostics,omitempty"`
}

// CompilerReport is frozen compiler evidence, not production certification.
type CompilerReport struct {
	Eligible    int              `json:"eligible_count"`
	Generated   int              `json:"generated_count"`
	Unsupported []DiagnosticCard `json:"unsupported"`
	Excluded    []DiagnosticCard `json:"excluded"`
}

// Record has an exact corpus identity and separate compiler/production evidence.
// LegacyCandidates are possible name/face associations, not exact admissions.
type Record struct {
	OracleID         string          `json:"oracle_id"`
	Name             string          `json:"name"`
	CompilerAccepted bool            `json:"compiler_accepted"`
	Production       *cards.Support  `json:"exact_production,omitempty"`
	LegacyCandidates []string        `json:"legacy_candidates,omitempty"`
	Diagnostics      json.RawMessage `json:"compiler_diagnostics,omitempty"`
	Blocker          string          `json:"gameplay_blocker,omitempty"`
}

// Inventory keeps production registrations and exact corpus evidence separate.
type Inventory struct {
	CorpusSHA256   string          `json:"corpus_sha256"`
	CompilerSHA256 string          `json:"compiler_report_sha256"`
	Eligible       int             `json:"eligible"`
	Accepted       int             `json:"compiler_accepted"`
	Production     []cards.Support `json:"production"`
	Corpus         []Record        `json:"corpus"`
	// Backlog is production absence, not a parser/lowering ranked roadmap.
	Backlog []Record `json:"production_backlog"`
}

// Load joins existing same-input files without running the compiler or parser.
func Load(corpusPath, compilerPath string, registry *cards.Registry) (Inventory, error) {
	corpusData, err := os.ReadFile(corpusPath)
	if err != nil {
		return Inventory{}, err
	}
	compilerData, err := os.ReadFile(compilerPath)
	if err != nil {
		return Inventory{}, err
	}
	var corpus []OracleCard
	var compiler CompilerReport
	if err := json.Unmarshal(corpusData, &corpus); err != nil {
		return Inventory{}, err
	}
	if err := json.Unmarshal(compilerData, &compiler); err != nil {
		return Inventory{}, err
	}
	return Merge(corpus, compiler, registry.Inventory(), fmt.Sprintf("%x", sha256.Sum256(corpusData)), fmt.Sprintf("%x", sha256.Sum256(compilerData)))
}

// Merge reconciles exact identities while retaining only candidate legacy links.
func Merge(corpus []OracleCard, compiler CompilerReport, production []cards.Support, corpusHash, compilerHash string) (Inventory, error) {
	out := Inventory{CorpusSHA256: corpusHash, CompilerSHA256: compilerHash, Eligible: compiler.Eligible, Accepted: compiler.Generated, Production: production}
	excluded := make(map[string]bool)
	rejected := make(map[string]DiagnosticCard)
	byName := make(map[string][]string)
	loadableLegacy := make(map[string]cards.Support)
	exact := make(map[string]cards.Support)
	quarantined := make(map[string]string)
	for _, exclusion := range cards.CompilerExclusions() {
		quarantined[exclusion.OracleID] = exclusion.Blocker
	}
	for _, card := range compiler.Excluded {
		excluded[card.OracleID] = true
	}
	for _, card := range compiler.Unsupported {
		rejected[card.OracleID] = card
	}
	for _, support := range production {
		if support.OracleID != "" {
			exact[support.OracleID] = support
		} else {
			byName[support.Name] = append(byName[support.Name], support.Key)
			loadableLegacy[support.Key] = support
		}
	}
	seen := make(map[string]bool)
	accepted := 0
	for _, card := range corpus {
		if carddata.DisownedCard(carddata.ScryfallCard{Name: card.Name}) {
			continue
		}
		if seen[card.OracleID] || card.OracleID == "" {
			return Inventory{}, fmt.Errorf("ambiguous corpus identity %q", card.OracleID)
		}
		seen[card.OracleID] = true
		if excluded[card.OracleID] {
			continue
		}
		diagnostic, rejected := rejected[card.OracleID]
		record := Record{OracleID: card.OracleID, Name: card.Name, CompilerAccepted: !rejected,
			Diagnostics: diagnostic.Diagnostics, Blocker: quarantined[card.OracleID]}
		if record.CompilerAccepted {
			accepted++
		}
		if support, exists := exact[card.OracleID]; exists {
			if support.Name != card.Name {
				return Inventory{}, fmt.Errorf("authored identity %s name mismatch: %q != %q", card.OracleID, support.Name, card.Name)
			}
			record.Production = &support
			delete(exact, card.OracleID)
		}
		record.LegacyCandidates = append(record.LegacyCandidates, byName[card.Name]...)
		for _, face := range card.Faces {
			for _, key := range byName[face.Name] {
				if !slices.Contains(record.LegacyCandidates, key) {
					record.LegacyCandidates = append(record.LegacyCandidates, key)
				}
			}
		}
		out.Corpus = append(out.Corpus, record)
		legacyLoadable := slices.ContainsFunc(record.LegacyCandidates, func(key string) bool {
			support := loadableLegacy[key]
			if !support.Loadable || card.Layout == "token" || card.Layout == "double_faced_token" {
				return false
			}
			if card.Name == support.Name {
				return true
			}
			return len(card.Faces) > 1 && !slices.ContainsFunc(card.Faces, func(face OracleFace) bool {
				return !slices.Contains(support.Faces, face.Name)
			})
		})
		if (record.Production == nil || !record.Production.Loadable) && !legacyLoadable {
			out.Backlog = append(out.Backlog, record)
		}
	}
	if len(exact) != 0 {
		return Inventory{}, fmt.Errorf("authored registrations absent from eligible corpus: %v", exact)
	}
	if len(out.Corpus) != compiler.Eligible || accepted != compiler.Generated {
		return Inventory{}, fmt.Errorf("compiler/corpus evidence does not reconcile: eligible %d/%d accepted %d/%d", len(out.Corpus), compiler.Eligible, accepted, compiler.Generated)
	}
	slices.SortFunc(out.Corpus, func(a, b Record) int {
		if a.Name < b.Name {
			return -1
		}
		if a.Name > b.Name {
			return 1
		}
		return 0
	})
	return out, nil
}
