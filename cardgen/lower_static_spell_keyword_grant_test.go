package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// staticGrantRuleEffect returns the lone RuleEffect carried by a face's single
// static ability, failing the test if the shape is unexpected.
func staticGrantRuleEffect(t *testing.T, face loweredFaceAbilities) game.RuleEffect {
	t.Helper()
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 1 {
		t.Fatalf("rule effects = %#v, want one", effects)
	}
	return effects[0]
}

// TestLowerStaticSpellKeywordGrant proves the static "[<filter>] spells you cast
// have <keyword>." family (Inspiring Statuary, Ironheart, Clever Champion,
// Caetus) lowers to a controller-scoped RuleEffectGrantSpellKeyword whose card
// selection encodes the card-type filter and whose granted keyword is the parsed
// cost-affecting keyword.
func TestLowerStaticSpellKeywordGrant(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		excluded   []types.Card
		keyword    game.Keyword
	}{
		"nonartifact improvise": {
			oracleText: "Nonartifact spells you cast have improvise.",
			excluded:   []types.Card{types.Artifact},
			keyword:    game.Improvise,
		},
		"noncreature improvise": {
			oracleText: "Noncreature spells you cast have improvise.",
			excluded:   []types.Card{types.Creature},
			keyword:    game.Improvise,
		},
		"noncreature convoke": {
			oracleText: "Noncreature spells you cast have convoke.",
			excluded:   []types.Card{types.Creature},
			keyword:    game.Convoke,
		},
		"unfiltered improvise": {
			oracleText: "Spells you cast have improvise.",
			excluded:   nil,
			keyword:    game.Improvise,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Statuary",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			effect := staticGrantRuleEffect(t, face)
			if effect.Kind != game.RuleEffectGrantSpellKeyword {
				t.Fatalf("kind = %v, want RuleEffectGrantSpellKeyword", effect.Kind)
			}
			if effect.AffectedController != game.ControllerYou {
				t.Fatalf("affected controller = %v, want ControllerYou", effect.AffectedController)
			}
			if effect.GrantedKeyword != test.keyword {
				t.Fatalf("granted keyword = %v, want %v", effect.GrantedKeyword, test.keyword)
			}
			if effect.AppliesToNextSpellOnly {
				t.Fatal("static grant must not be limited to the next spell")
			}
			if !slices.Equal(effect.CardSelection.ExcludedTypes, test.excluded) {
				t.Fatalf("excluded types = %v, want %v", effect.CardSelection.ExcludedTypes, test.excluded)
			}
		})
	}
}

// TestLowerStaticSpellKeywordGrantFailsClosed proves grants the payment
// machinery cannot honor are reported unsupported rather than generating a rule
// effect the runtime would silently ignore: a non-cost keyword and an
// unrecognized card-type filter both fail closed.
func TestLowerStaticSpellKeywordGrantFailsClosed(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"non cost keyword":   "Spells you cast have flying.",
		"unsupported filter": "Nonland spells you cast have improvise.",
	}
	for name, oracleText := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Test Fail Statuary",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: oracleText,
			})
			if !face.empty() {
				t.Fatalf("expected no partial ability, got %#v", face)
			}
		})
	}
}

// TestGenerateInspiringStatuarySource proves the static grant renders to
// executable CardDef source that names the rule-effect kind, the granted
// keyword, the controller scope, and the nonartifact card selection.
func TestGenerateInspiringStatuarySource(t *testing.T) {
	t.Parallel()
	source := generateGrantSource(t, &ScryfallCard{
		Name:       "Inspiring Statuary",
		Layout:     "normal",
		ManaCost:   "{3}",
		TypeLine:   "Artifact",
		OracleText: "Nonartifact spells you cast have improvise.",
	})
	for _, wanted := range []string{
		"Kind:               game.RuleEffectGrantSpellKeyword",
		"GrantedKeyword:     game.Improvise",
		"AffectedController: game.ControllerYou",
		"ExcludedTypes: []types.Card{types.Artifact}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateIronheartComposesNativeAndGrantedImprovise proves a card that both
// carries Improvise natively and grants it to other spells (Ironheart, Clever
// Champion) renders the native keyword body and the grant as independent static
// abilities, so the two never double-apply.
func TestGenerateIronheartComposesNativeAndGrantedImprovise(t *testing.T) {
	t.Parallel()
	source := generateGrantSource(t, &ScryfallCard{
		Name:       "Ironheart, Clever Champion",
		Layout:     "normal",
		ManaCost:   "{4}{U}",
		TypeLine:   "Legendary Artifact Creature — Human Hero",
		OracleText: "Improvise (Your artifacts can help cast this spell.)\nFlying\nNoncreature spells you cast have improvise.",
		Colors:     []string{"U"},
		Power:      new("3"),
		Toughness:  new("4"),
	})
	for _, wanted := range []string{
		"game.ImproviseStaticBody",
		"Kind:               game.RuleEffectGrantSpellKeyword",
		"GrantedKeyword:     game.Improvise",
		"ExcludedTypes: []types.Card{types.Creature}",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// generateGrantSource generates executable CardDef source for card, failing the
// test on any generation error or diagnostic.
func generateGrantSource(t *testing.T, card *ScryfallCard) string {
	t.Helper()
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	return source
}
