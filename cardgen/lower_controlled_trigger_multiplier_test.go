package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
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

// controlledMultiplierRuleEffect pulls the single
// RuleEffectAdditionalTriggerForControlledPermanent out of a one-static,
// one-rule-effect face, failing the test on any other shape. Unlike
// controlledMultiplierSelection it returns the whole rule effect so callers can
// assert on non-selection fields such as TriggerCauseCastOrCopyInstantSorcery.
func controlledMultiplierRuleEffect(t *testing.T, face loweredFaceAbilities) game.RuleEffect {
	t.Helper()
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v", face.StaticAbilities)
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 1 || effects[0].Kind != game.RuleEffectAdditionalTriggerForControlledPermanent {
		t.Fatalf("rule effects = %#v", effects)
	}
	return effects[0]
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
		"another subtype excludes self": {
			oracleText: "If a triggered ability of another Elemental you control triggers, it triggers an additional time.",
			want: game.Selection{
				SubtypesAny:   []types.Sub{types.Sub("Elemental")},
				ExcludeSource: true,
			},
		},
		"another card type with that-ability tail": {
			oracleText: "If a triggered ability of another creature you control triggers, that ability triggers an additional time.",
			want: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				ExcludeSource: true,
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
				!equalSubSlice(got.SubtypesAny, tc.want.SubtypesAny) ||
				got.ExcludeSource != tc.want.ExcludeSource {
				t.Fatalf("selection = %#v, want %#v", got, tc.want)
			}
		})
	}
}

// TestLowerControlledTriggerMultiplierDisjunction verifies the "or"-joined filter
// "a Shaman or another Wizard you control" (Harmonic Prodigy) lowers to an AnyOf
// of two branch selections, the second carrying ExcludeSource for "another".
func TestLowerControlledTriggerMultiplierDisjunction(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test disjunction",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Wizard",
		OracleText: "If a triggered ability of a Shaman or another Wizard you control triggers, that ability triggers an additional time.",
	})
	got := controlledMultiplierSelection(t, face)
	if len(got.AnyOf) != 2 {
		t.Fatalf("AnyOf = %#v, want two branches", got.AnyOf)
	}
	shaman := got.AnyOf[0]
	if !equalSubSlice(shaman.SubtypesAny, []types.Sub{types.Sub("Shaman")}) || shaman.ExcludeSource {
		t.Fatalf("first branch = %#v, want bare Shaman", shaman)
	}
	wizard := got.AnyOf[1]
	if !equalSubSlice(wizard.SubtypesAny, []types.Sub{types.Sub("Wizard")}) || !wizard.ExcludeSource {
		t.Fatalf("second branch = %#v, want Wizard excluding self", wizard)
	}
}

// TestLowerControlledTriggerMultiplierPower verifies Delney, Streetwise Lookout's
// "a creature you control with power 2 or less" source filter lowers to a
// power-bounded creature Selection. Delney itself is power 2 and the authoritative
// Oracle text reads "a creature" (not "another"), so the filter must NOT set
// ExcludeSource: Delney doubles its own controlled triggered abilities too.
func TestLowerControlledTriggerMultiplierPower(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		oracleText string
		want       compare.Int
	}{
		"power 2 or less (Delney)": {
			oracleText: "If a triggered ability of a creature you control with power 2 or less triggers, that ability triggers an additional time.",
			want:       compare.Int{Op: compare.LessOrEqual, Value: 2},
		},
		"power 4 or greater": {
			oracleText: "If a triggered ability of a creature you control with power 4 or greater triggers, that ability triggers an additional time.",
			want:       compare.Int{Op: compare.GreaterOrEqual, Value: 4},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Power Filter " + name,
				Layout:     "normal",
				TypeLine:   "Legendary Enchantment",
				OracleText: tc.oracleText,
			})
			got := controlledMultiplierSelection(t, face)
			if got.ExcludeSource {
				t.Fatalf("ExcludeSource = true, want false (self-inclusion): %#v", got)
			}
			if !equalCardSlice(got.RequiredTypes, []types.Card{types.Creature}) {
				t.Fatalf("RequiredTypes = %#v, want [Creature]", got.RequiredTypes)
			}
			if !got.Power.Exists || got.Power.Val != tc.want {
				t.Fatalf("Power = %#v, want set %#v", got.Power, tc.want)
			}
		})
	}
}

// TestLowerControlledTriggerMultiplierMagecraft verifies Veyran, Voice of
// Duality's "If you casting or copying an instant or sorcery spell causes a
// triggered ability of a permanent you control to trigger" clause lowers to the
// causal magecraft doubler: TriggerCauseCastOrCopyInstantSorcery set, with an
// empty source selection ("a permanent", no exclusion) and no branch or power
// filter. Veyran itself is a permanent you control, so it doubles its own
// magecraft-caused triggered abilities.
func TestLowerControlledTriggerMultiplierMagecraft(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Magecraft Doubler",
		Layout:     "normal",
		TypeLine:   "Legendary Enchantment",
		OracleText: "If you casting or copying an instant or sorcery spell causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
	})
	effect := controlledMultiplierRuleEffect(t, face)
	if !effect.TriggerCauseCastOrCopyInstantSorcery {
		t.Fatalf("TriggerCauseCastOrCopyInstantSorcery = false, want true: %#v", effect)
	}
	if !effect.AffectedSelection.Empty() {
		t.Fatalf("AffectedSelection = %#v, want empty (any permanent you control)", effect.AffectedSelection)
	}
}

func TestLowerControlledTriggerMultiplierFailsClosed(t *testing.T) {
	t.Parallel()
	for name, oracleText := range map[string]string{
		"bare supertype no noun": "If a triggered ability of a legendary permanent you control triggers, that ability triggers an additional time.",
		"branch missing article": "If a triggered ability of a Shaman or Wizard you control triggers, that ability triggers an additional time.",
		"bare another supertype": "If a triggered ability of another legendary you control triggers, that ability triggers an additional time.",
		"magecraft wrong spell":  "If you casting or copying an artifact spell causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
		"power wrong stat":       "If a triggered ability of a creature you control with toughness 2 or less triggers, that ability triggers an additional time.",
		"power wrong comparator": "If a triggered ability of a creature you control with power 2 or fewer triggers, that ability triggers an additional time.",
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
		"harmonic prodigy": {
			card: &ScryfallCard{
				Name:       "Harmonic Prodigy",
				Layout:     "normal",
				ManaCost:   "{1}{R}",
				TypeLine:   "Legendary Creature — Human Wizard",
				OracleText: "Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)\nIf a triggered ability of a Shaman or another Wizard you control triggers, that ability triggers an additional time.",
				Power:      new("1"),
				Toughness:  new("3"),
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForControlledPermanent",
				`AnyOf: []game.Selection{game.Selection{SubtypesAny: []types.Sub{types.Sub("Shaman")}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Wizard")}, ExcludeSource: true}}`,
			},
		},
		"twinflame travelers": {
			card: &ScryfallCard{
				Name:       "Twinflame Travelers",
				Layout:     "normal",
				ManaCost:   "{2}{U}{R}",
				TypeLine:   "Creature — Elemental Sorcerer",
				OracleText: "Flying\nIf a triggered ability of another Elemental you control triggers, it triggers an additional time.",
				Power:      new("3"),
				Toughness:  new("3"),
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForControlledPermanent",
				`SubtypesAny: []types.Sub{types.Sub("Elemental")}, ExcludeSource: true`,
			},
		},
		"veyran voice of duality": {
			card: &ScryfallCard{
				Name:       "Veyran, Voice of Duality",
				Layout:     "normal",
				ManaCost:   "{1}{U}{R}",
				TypeLine:   "Legendary Creature — Efreet Wizard",
				OracleText: "Magecraft — Whenever you cast or copy an instant or sorcery spell, Veyran, Voice of Duality gets +1/+1 until end of turn.\nIf you casting or copying an instant or sorcery spell causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
				Power:      new("2"),
				Toughness:  new("2"),
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForControlledPermanent",
				"TriggerCauseCastOrCopyInstantSorcery: true",
				"game.EventSpellCast",
				"MatchSpellCopy: true",
			},
		},
		// Delney's full card is not curatable (its "can't be blocked by creatures
		// with power 3 or greater" evasion static does not parse), but its
		// authoritative trigger sentence lowers on its own. It reads "a creature"
		// (self-inclusion), so no ExcludeSource, and carries the power-2 bound.
		"delney trigger clause": {
			card: &ScryfallCard{
				Name:       "Delney Trigger Clause",
				Layout:     "normal",
				TypeLine:   "Legendary Creature — Human Scout",
				OracleText: "If a triggered ability of a creature you control with power 2 or less triggers, that ability triggers an additional time.",
				Power:      new("2"),
				Toughness:  new("2"),
			},
			wants: []string{
				"game.RuleEffectAdditionalTriggerForControlledPermanent",
				"RequiredTypes: []types.Card{types.Creature}",
				"Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})",
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
	return slices.Equal(a, b)
}

func equalSuperSlice(a, b []types.Super) bool {
	return slices.Equal(a, b)
}

func equalSubSlice(a, b []types.Sub) bool {
	return slices.Equal(a, b)
}
