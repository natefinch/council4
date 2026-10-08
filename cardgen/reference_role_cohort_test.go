package cardgen

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func referenceRoleOriginals(t *testing.T) map[string]ScryfallCard {
	t.Helper()
	data, err := os.ReadFile("testdata/reference-role-originals.json")
	if err != nil {
		t.Fatal(err)
	}
	var cards []ScryfallCard
	if err := json.Unmarshal(data, &cards); err != nil {
		t.Fatal(err)
	}
	if len(cards) != 618 {
		t.Fatalf("original cohort count = %d, want 618", len(cards))
	}
	result := make(map[string]ScryfallCard, len(cards))
	for _, card := range cards {
		if _, duplicate := result[card.Name]; duplicate {
			t.Fatalf("duplicate original identity %q", card.Name)
		}
		result[card.Name] = card
	}
	return result
}

func TestReferenceRoleCompleteOriginalLossCohort(t *testing.T) {
	cards := referenceRoleOriginals(t)
	data, err := os.ReadFile("testdata/reference-role-loss-identities.json")
	if err != nil {
		t.Fatal(err)
	}
	var identities [][4]string
	if err := json.Unmarshal(data, &identities); err != nil {
		t.Fatal(err)
	}
	if len(identities) != 615 {
		t.Fatalf("loss identity count = %d, want 615", len(identities))
	}
	for _, identity := range identities {
		t.Run(identity[2], func(t *testing.T) {
			card, exists := cards[identity[2]]
			if !exists || card.ID != identity[0] || card.OracleID != identity[1] || card.Layout != identity[3] {
				t.Fatal("complete original identity changed")
			}
			defs, diagnostics, err := CompileCardDefs(&card)
			if err != nil || len(diagnostics) != 0 || len(defs) == 0 {
				logReferenceRoleProofs(t, card)
				t.Fatalf("whole-card loss remains: err=%v diagnostics=%#v definitions=%d", err, diagnostics, len(defs))
			}

		})
	}
}

func logReferenceRoleProofs(t *testing.T, card ScryfallCard) {
	t.Helper()
	logBody := func(name, text, typeLine string) {
		document, _ := parser.Parse(text, parser.Context{CardName: name,
			InstantOrSorcery: strings.Contains(typeLine, "Instant") || strings.Contains(typeLine, "Sorcery")})
		compilation, _ := compiler.Compile(document, compiler.Context{})
		var logContent func(compiler.AbilityContent)
		logContent = func(content compiler.AbilityContent) {
			encodedConditions, err := json.Marshal(content.Conditions)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("typed conditions: %s", encodedConditions)
			for _, reference := range content.References {
				if !reference.SubjectSupported() {
					t.Logf("unavailable proof: %+v", reference)
				}
			}
			for _, effect := range content.Effects {
				encoded, err := json.Marshal(effect)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("typed effect: %s", encoded)
				t.Logf("effect kind=%v clause=%d targetCount=%d refs=%v exact=%v from=%v to=%v amountRef=%d",
					effect.Kind, effect.ClauseID, len(effect.Targets), effect.References, effect.Exact,
					effect.FromZone, effect.ToZone, effect.Amount.ReferenceNodeID)
			}
			for _, mode := range content.Modes {
				logContent(mode.Content)
			}
		}
		for _, ability := range compilation.Abilities {
			if ability.Trigger != nil {
				t.Logf("typed trigger: %+v", *ability.Trigger)
			}
			logContent(ability.Content)
			for _, effect := range ability.Content.Effects {
				if effect.DelayedTriggerAbility != nil {
					inner, _ := effect.DelayedTriggerAbility.Inner()
					nested, _ := compiler.Compile(inner, compiler.Context{})
					for _, child := range nested.Abilities {
						if child.Trigger != nil {
							t.Logf("typed nested trigger: %+v", *child.Trigger)
						}
						logContent(child.Content)
					}
				}
			}
		}
	}
	if len(card.CardFaces) == 0 {
		logBody(card.Name, card.OracleText, card.TypeLine)
	}
	for _, face := range card.CardFaces {
		logBody(face.Name, face.OracleText, face.TypeLine)
	}
}

func TestReferenceRoleCompoundPaymentRefusesWholeCard(t *testing.T) {
	card := referenceRoleOriginals(t)["Lamplight Phoenix"]
	defs, diagnostics, err := CompileCardDefs(&card)
	if err != nil {
		t.Fatal(err)
	}
	if len(defs) != 0 || len(diagnostics) == 0 {
		t.Fatalf("unmodeled compound exile/evidence payment partially admitted: definitions=%d diagnostics=%#v", len(defs), diagnostics)
	}
}
