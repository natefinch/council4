package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// controlledMultiplierSelection pulls the single
// RuleEffectAdditionalTriggerForControlledPermanent out of a one-static,
// one-rule-effect face, failing the test on any other shape.
func controlledMultiplierSelection(t *testing.T, face loweredFaceAbilities) game.Selection {
	t.Helper()
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v", face.StaticAbilities)
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 1 || effects[0].Kind != game.RuleEffectAdditionalTriggerForControlledPermanent {
		t.Fatalf("rule effects = %#v", effects)
	}
	return effects[0].AffectedSelection
}

func TestLowerControlledTriggerMultiplierFilters(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		oracleText string
		want       game.Selection
	}{
		"legendary creature": {
			oracleText: "If a triggered ability of a legendary creature you control triggers, that ability triggers an additional time.",
			want: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Supertypes:    []types.Super{types.Legendary},
			},
		},
		"bare subtype": {
			oracleText: "If a triggered ability of an Ally you control triggers, that ability triggers an additional time.",
			want:       game.Selection{SubtypesAny: []types.Sub{types.Sub("Ally")}},
		},
		"subtype and card type": {
			oracleText: "If a triggered ability of a Ninja creature you control triggers, that ability triggers an additional time.",
			want: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				SubtypesAny:   []types.Sub{types.Sub("Ninja")},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test " + name,
				Layout:     "normal",
				TypeLine:   "Legendary Enchantment",
				OracleText: tc.oracleText,
			})
			got := controlledMultiplierSelection(t, face)
			if got.AnyOf != nil {
				t.Fatalf("unexpected AnyOf in %#v", got)
			}
			if !equalCardSlice(got.RequiredTypes, tc.want.RequiredTypes) ||
				!equalSuperSlice(got.Supertypes, tc.want.Supertypes) ||
				!equalSubSlice(got.SubtypesAny, tc.want.SubtypesAny) {
				t.Fatalf("selection = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestLowerControlledTriggerMultiplierFailsClosed(t *testing.T) {
	t.Parallel()
	for name, oracleText := range map[string]string{
		"bare supertype no noun": "If a triggered ability of a legendary permanent you control triggers, that ability triggers an additional time.",
		"or-joined subtypes":     "If a triggered ability of a Shaman or another Wizard you control triggers, that ability triggers an additional time.",
		"another qualifier":      "If a triggered ability of another creature you control triggers, that ability triggers an additional time.",
		"it-triggers tail":       "If a triggered ability of a legendary creature you control triggers, it triggers an additional time.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Near Miss " + name,
				Layout:     "normal",
				TypeLine:   "Legendary Enchantment",
				OracleText: oracleText,
			})
		})
	}
}

func TestGenerateExecutableControlledTriggerMultiplierCards(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		card  *ScryfallCard
		wants []string
	}{
		"annie joins up": {
			card: &ScryfallCard{
				Name:       "Annie Joins Up",
				Layout:     "normal",
				ManaCost:   "{1}{R}{G}{W}",
				TypeLine:   "Legendary Enchantment",
				OracleText: "When Annie Joins Up enters, it deals 5 damage to target creature or planeswalker an opponent controls.\nIf a triggered ability of a legendary creature you control triggers, that ability triggers an additional time.",
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForControlledPermanent",
				"Supertypes: []types.Super{types.Legendary}",
				"RequiredTypes: []types.Card{types.Creature}",
			},
		},
		"katara the fearless": {
			card: &ScryfallCard{
				Name:       "Katara, the Fearless",
				Layout:     "normal",
				ManaCost:   "{2}{R}{W}",
				TypeLine:   "Legendary Creature — Human Warrior Ally",
				OracleText: "If a triggered ability of an Ally you control triggers, that ability triggers an additional time.",
				Power:      new("3"),
				Toughness:  new("3"),
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForControlledPermanent",
				`SubtypesAny: []types.Sub{types.Sub("Ally")}`,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(tc.card, "p")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range tc.wants {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

func equalCardSlice(a, b []types.Card) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalSuperSlice(a, b []types.Super) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalSubSlice(a, b []types.Sub) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
