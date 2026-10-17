package cards_test

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
)

const historicalGenericVerification = "mtg/rules named-card scenarios; mixed production acceptance (#3278)"

func TestAuthoredRegistryRequiresExplicitVerification(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name         string
		verification string
		valid        bool
	}{
		{name: "missing"},
		{name: "whitespace", verification: " \t\r\n\u00a0"},
		{name: "historical generic", verification: historicalGenericVerification},
		{name: "padded historical generic", verification: "\t" + historicalGenericVerification + "\n"},
		{
			name:         "explicit fixture evidence",
			verification: "mtg/cards/verification_registry_test.go: TestAuthoredRegistryRequiresExplicitVerification (test fixture)",
			valid:        true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			constructed := false
			entry := cardset.Entry{
				Name: "Verification fixture", OracleID: "fixture/verification",
				Verification: test.verification,
				New: func() *game.CardDef {
					constructed = true
					return &game.CardDef{CardFace: game.CardFace{Name: "Verification fixture"}}
				},
			}
			registry, err := cards.NewProgramRegistry([]cardset.Entry{entry})
			if constructed {
				t.Fatal("verification admission eagerly constructed the card")
			}
			if !test.valid {
				if err == nil || !strings.Contains(err.Error(), "explicit card-specific verification evidence") || !strings.Contains(err.Error(), entry.Name) {
					t.Fatalf("verification diagnostic = %v", err)
				}
				if registry != nil {
					t.Fatal("failed admission returned a usable registry")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			def, err := registry.LookupOracleID(entry.OracleID)
			if err != nil || def == nil || def.Name != entry.Name || !constructed {
				t.Fatalf("explicit verified fixture lookup = %v, %v", def, err)
			}
		})
	}
}

func TestAuthoredStaticAndProgramEntriesKeepExplicitVerification(t *testing.T) {
	t.Parallel()
	for _, entry := range []cardset.Entry{authored.BloodMoon, authored.HiveMind} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			registry, err := cards.NewProgramRegistry([]cardset.Entry{entry})
			if err != nil {
				t.Fatal(err)
			}
			def, err := registry.LookupOracleID(entry.OracleID)
			if err != nil || def == nil || def.Name != entry.Name {
				t.Fatalf("existing authored lookup = %v, %v", def, err)
			}
			evidence := registry.SupportFor(entry.Name)
			if len(evidence) != 1 || !evidence[0].Loadable || !evidence[0].BehaviorVerified || evidence[0].Verification != entry.Verification {
				t.Fatalf("explicit authored evidence was lost: %+v", evidence)
			}
		})
	}
}

func TestLegacyNameOnlyRegistryDoesNotRequireVerification(t *testing.T) {
	t.Parallel()
	def := &game.CardDef{CardFace: game.CardFace{Name: "Legacy fixture"}}
	registry, err := cards.NewProgramRegistry([]cardset.Entry{{
		Name: def.Name, New: func() *game.CardDef { return def },
	}})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := registry.LookupChecked(def.Name); err != nil || got != def {
		t.Fatalf("legacy lookup = %v, %v", got, err)
	}
	evidence := registry.SupportFor(def.Name)
	if len(evidence) != 1 || !evidence[0].Loadable || evidence[0].OracleID != "" || evidence[0].BehaviorVerified {
		t.Fatalf("legacy registration was promoted or rejected: %+v", evidence)
	}
}
