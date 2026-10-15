package cards_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/tokens"
)

type installedCompilerCard struct {
	OracleID    string `json:"oracle_id"`
	Name        string `json:"name"`
	Layout      string `json:"layout"`
	Path        string `json:"path"`
	Constructor string `json:"constructor"`
	SHA256      string `json:"sha256"`
}

func TestCompletePinnedCompilerSourcesAreInstalled(t *testing.T) {
	data, err := os.ReadFile("../../testdata/card-preservation/installed-compiler.json")
	if err != nil {
		t.Fatal(err)
	}
	var installation struct {
		CorpusSHA256 string                  `json:"corpus_sha256"`
		Cards        []installedCompilerCard `json:"cards"`
	}
	if err := json.Unmarshal(data, &installation); err != nil {
		t.Fatal(err)
	}
	if installation.CorpusSHA256 != "6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4" ||
		len(installation.Cards) != 18700 {
		t.Fatal("pinned compiler coverage changed without reconciliation")
	}
	registry := cards.NewDefaultRegistry()
	exclusions := make(map[string]cards.CompilerExclusion)
	for _, exclusion := range cards.CompilerExclusions() {
		exclusions[exclusion.OracleID] = exclusion
	}
	tokenNames := make(map[string]bool)
	for _, token := range tokens.Cards {
		tokenNames[token.Name] = true
	}
	seen := make(map[string]bool)
	deckCount, tokenCount := 0, 0
	for _, card := range installation.Cards {
		if card.OracleID == "" || card.Constructor == "" || seen[card.OracleID] || !filepath.IsLocal(card.Path) {
			t.Fatalf("missing, duplicate or invalid compiled identity: %+v", card)
		}
		seen[card.OracleID] = true
		source, err := os.ReadFile(card.Path)
		if err != nil {
			t.Fatalf("%s source: %v", card.Name, err)
		}
		if got := fmt.Sprintf("%x", sha256.Sum256(source)); got != card.SHA256 {
			t.Fatalf("%s source changed without reconciling installed coverage", card.Name)
		}
		if card.Layout == "token" || card.Layout == "double_faced_token" {
			tokenCount++
			name, _, _ := strings.Cut(card.Name, " // ")
			if !strings.HasPrefix(card.Path, "tokens/") || !tokenNames[name] {
				t.Fatalf("token %s is not in the token catalog", card.Name)
			}
			continue
		}
		deckCount++
		if exclusion, blocked := exclusions[card.OracleID]; blocked {
			if exclusion.Name != card.Name || exclusion.Blocker == "" || len(registry.LookupAll(card.Name)) != 0 {
				t.Fatalf("known broken compiled card was promoted: %s", card.Name)
			}
			continue
		}
		name := card.Name
		if registry.Lookup(name) == nil {
			name, _, _ = strings.Cut(name, " // ")
		}
		if def, err := registry.LookupChecked(name); err != nil || def == nil {
			t.Fatalf("compiler-accepted card %s is absent from production: %v", card.Name, err)
		}
	}
	if deckCount != 17946 || tokenCount != 754 {
		t.Fatalf("compiler population: %d deck cards, %d tokens", deckCount, tokenCount)
	}
}

func TestAuthoredImplementationsSupersedeTheirGeneratedCopies(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	replacements := 0
	for _, entry := range authored.Cards {
		if !entry.ReplaceLegacy {
			continue
		}
		replacements++
		def, err := registry.LookupOracleID(entry.OracleID)
		if err != nil || def == nil || len(registry.LookupAll(entry.Name)) != 1 {
			t.Fatalf("authored implementation lost precedence for %s: %v", entry.Name, err)
		}
		support := registry.SupportFor(entry.Name)
		if len(support) != 1 || !support[0].BehaviorVerified || support[0].OracleID != entry.OracleID {
			t.Fatalf("generated copy replaced authored evidence for %s", entry.Name)
		}
	}
	if replacements != 24 {
		t.Fatalf("authored/compiler overlap = %d, want 24", replacements)
	}
}
