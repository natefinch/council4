package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerNextSpellKeywordGrant proves the one-shot "The next spell you cast
// this turn has <keyword>." grant (Archway of Innovation, Wand of the Worldsoul)
// and the all-spells "Spells you cast this turn have <keyword>." form both lower
// to a turn-scoped ApplyRule carrying a controller-scoped
// RuleEffectGrantSpellKeyword; only the next-spell form limits the grant to the
// single next spell the controller casts.
func TestLowerNextSpellKeywordGrant(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		keyword    game.Keyword
		nextOnly   bool
	}{
		"next spell improvise": {
			oracleText: "{T}: The next spell you cast this turn has improvise.",
			keyword:    game.Improvise,
			nextOnly:   true,
		},
		"next spell convoke": {
			oracleText: "{T}: The next spell you cast this turn has convoke.",
			keyword:    game.Convoke,
			nextOnly:   true,
		},
		"all spells improvise": {
			oracleText: "{T}: Spells you cast this turn have improvise.",
			keyword:    game.Improvise,
			nextOnly:   false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Archway",
				Layout:     "normal",
				TypeLine:   "Land",
				OracleText: test.oracleText,
			})
			apply := applyRuleFromActivated(t, face)
			if apply.Duration != game.DurationThisTurn {
				t.Fatalf("duration = %v, want DurationThisTurn", apply.Duration)
			}
			if len(apply.RuleEffects) != 1 {
				t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
			}
			effect := apply.RuleEffects[0]
			if effect.Kind != game.RuleEffectGrantSpellKeyword {
				t.Fatalf("kind = %v, want RuleEffectGrantSpellKeyword", effect.Kind)
			}
			if effect.AffectedController != game.ControllerYou {
				t.Fatalf("affected controller = %v, want ControllerYou", effect.AffectedController)
			}
			if effect.GrantedKeyword != test.keyword {
				t.Fatalf("granted keyword = %v, want %v", effect.GrantedKeyword, test.keyword)
			}
			if effect.AppliesToNextSpellOnly != test.nextOnly {
				t.Fatalf("applies to next spell only = %v, want %v", effect.AppliesToNextSpellOnly, test.nextOnly)
			}
		})
	}
}

// TestLowerNextSpellKeywordGrantFailsClosed proves the one-shot grant reports a
// keyword the payment machinery cannot honor as unsupported rather than emitting
// a grant the runtime would ignore.
func TestLowerNextSpellKeywordGrantFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Fail Archway",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: The next spell you cast this turn has flying.",
	})
	if !face.empty() {
		t.Fatalf("expected no partial ability, got %#v", face)
	}
}

// TestGenerateArchwayOfInnovationSource proves the one-shot grant renders to
// executable CardDef source that names the rule-effect kind, the granted
// keyword, the controller scope, the next-spell limiter, and the turn duration.
func TestGenerateArchwayOfInnovationSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Archway of Innovation",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {U}.\n{U}, {T}: The next spell you cast this turn has improvise.",
		Colors:     []string{"U"},
	}, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:                   game.RuleEffectGrantSpellKeyword",
		"GrantedKeyword:         game.Improvise",
		"AffectedController:     game.ControllerYou",
		"AppliesToNextSpellOnly: true",
		"Duration: game.DurationThisTurn",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
