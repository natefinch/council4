package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// enchantedDynamicPTContinuousEffects lowers an Aura whose enchant ability is
// followed by a dynamic "Enchanted creature gets +N/+N for each <count>" buff
// and returns the continuous effects of the buff's static ability.
func enchantedDynamicPTContinuousEffects(t *testing.T, oracleText string) []game.ContinuousEffect {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dynamic Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: oracleText,
	})
	if len(face.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %d, want 2 (enchant + buff)", len(face.StaticAbilities))
	}
	return face.StaticAbilities[1].Body.ContinuousEffects
}

func assertDynamicPTAnyTypes(t *testing.T, effect *game.ContinuousEffect, want ...types.Card) {
	t.Helper()
	if effect.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("layer = %v, want LayerPowerToughnessModify", effect.Layer)
	}
	if effect.Group.Domain() != game.GroupDomainAttachedObject {
		t.Fatalf("group domain = %v, want attached object", effect.Group.Domain())
	}
	if !effect.PowerDeltaDynamic.Exists || !effect.ToughnessDeltaDynamic.Exists {
		t.Fatalf("power/toughness deltas are not both dynamic: %+v", effect)
	}
	assertCountSelectorAnyTypes(t, "power", &effect.PowerDeltaDynamic.Val, want)
	assertCountSelectorAnyTypes(t, "toughness", &effect.ToughnessDeltaDynamic.Val, want)
}

func assertCountSelectorAnyTypes(t *testing.T, name string, amount *game.DynamicAmount, want []types.Card) {
	t.Helper()
	if amount.Kind != game.DynamicAmountCountSelector || amount.Multiplier != 1 {
		t.Fatalf("%s amount = %+v, want count-selector multiplier 1", name, amount)
	}
	selection := amount.Group.Selection()
	if selection.Controller != game.ControllerYou {
		t.Fatalf("%s controller = %v, want you", name, selection.Controller)
	}
	if len(selection.RequiredTypesAny) != len(want) {
		t.Fatalf("%s required-types-any = %+v, want %v", name, selection.RequiredTypesAny, want)
	}
	for i, card := range want {
		if selection.RequiredTypesAny[i] != card {
			t.Fatalf("%s required-types-any = %+v, want %v", name, selection.RequiredTypesAny, want)
		}
	}
}

// TestLowerDynamicAuraPTTypeUnion covers All That Glitters: the enchanted
// creature gets +1/+1 for each member of a card-type disjunction the
// controller owns, lowered to a count over Selection.RequiredTypesAny.
func TestLowerDynamicAuraPTTypeUnion(t *testing.T) {
	t.Parallel()
	effects := enchantedDynamicPTContinuousEffects(t,
		"Enchant creature\nEnchanted creature gets +1/+1 for each artifact and/or enchantment you control.")
	if len(effects) != 1 {
		t.Fatalf("continuous effects = %d, want 1", len(effects))
	}
	assertDynamicPTAnyTypes(t, &effects[0], types.Artifact, types.Enchantment)
}

// TestLowerDynamicAuraPTWithKeywordGrant covers Ethereal Armor: a dynamic
// +1/+1-for-each buff conjoined with a keyword grant must lower to two
// continuous effects (the dynamic P/T modify and the first-strike grant).
func TestLowerDynamicAuraPTWithKeywordGrant(t *testing.T) {
	t.Parallel()
	effects := enchantedDynamicPTContinuousEffects(t,
		"Enchant creature\nEnchanted creature gets +1/+1 for each enchantment you control and has first strike.")
	if len(effects) != 2 {
		t.Fatalf("continuous effects = %d, want 2 (dynamic P/T + keyword)", len(effects))
	}
	modify, grant := -1, -1
	for i := range effects {
		if effects[i].Layer == game.LayerPowerToughnessModify {
			modify = i
		}
		if effects[i].Layer == game.LayerAbility {
			grant = i
		}
	}
	if modify < 0 || grant < 0 {
		t.Fatalf("missing effects: modify=%d grant=%d", modify, grant)
	}
	assertDynamicPTSingleType(t, &effects[modify], types.Enchantment)
	if len(effects[grant].AddKeywords) != 1 || effects[grant].AddKeywords[0] != game.FirstStrike {
		t.Fatalf("granted keywords = %v, want [FirstStrike]", effects[grant].AddKeywords)
	}
}

func assertDynamicPTSingleType(t *testing.T, effect *game.ContinuousEffect, want types.Card) {
	t.Helper()
	if effect.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("layer = %v, want LayerPowerToughnessModify", effect.Layer)
	}
	if !effect.PowerDeltaDynamic.Exists || !effect.ToughnessDeltaDynamic.Exists {
		t.Fatalf("power/toughness deltas are not both dynamic: %+v", effect)
	}
	assertCountSelectorSingleType(t, "power", &effect.PowerDeltaDynamic.Val, want)
	assertCountSelectorSingleType(t, "toughness", &effect.ToughnessDeltaDynamic.Val, want)
}

func assertCountSelectorSingleType(t *testing.T, name string, amount *game.DynamicAmount, want types.Card) {
	t.Helper()
	selection := amount.Group.Selection()
	if len(selection.RequiredTypes) != 1 || selection.RequiredTypes[0] != want {
		t.Fatalf("%s required types = %+v, want [%v]", name, selection.RequiredTypes, want)
	}
	if selection.Controller != game.ControllerYou {
		t.Fatalf("%s controller = %v, want you", name, selection.Controller)
	}
}

// TestLowerDynamicAuraPTConjunctionFailsClosed guards the union/conjunction
// boundary: "artifact creature you control" is an intersection, not a card-type
// disjunction, so it must not be misread as a two-type union count and instead
// fails closed.
func TestLowerDynamicAuraPTConjunctionFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Dynamic Aura Conjunction",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\nEnchanted creature gets +1/+1 for each artifact creature you control.",
	})
}
