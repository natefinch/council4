package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func spellTypeEffectCard(effect ContinuousEffect, zoneOfFunction zone.Type) *CardDef {
	return &CardDef{CardFace: CardFace{
		Name:  "Spell Type Effect",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []StaticAbility{{
			ZoneOfFunction:    zoneOfFunction,
			ContinuousEffects: []ContinuousEffect{effect},
		}},
	}}
}

func creatureSpellTypeEffect() ContinuousEffect {
	return ContinuousEffect{
		Duration:       DurationPermanent,
		Group:          ObjectControlledGroup(SourcePermanentReference(), Selection{RequiredTypes: []types.Card{types.Creature}}),
		Layer:          LayerType,
		SpellSelection: &Selection{RequiredTypes: []types.Card{types.Creature}},
		AddTypes:       []types.Card{types.Artifact},
	}
}

// TestValidateCardDefAllowsSpellTypeEffect proves a battlefield type-changing
// static ability may extend to spells by type-line selection.
func TestValidateCardDefAllowsSpellTypeEffect(t *testing.T) {
	t.Parallel()
	if issues := ValidateCardDef(spellTypeEffectCard(creatureSpellTypeEffect(), zone.None)); len(issues) != 0 {
		t.Fatalf("issues = %+v, want none", issues)
	}
	effect := creatureSpellTypeEffect()
	effect.SpellSelection = &Selection{AnyOf: []Selection{
		{RequiredTypes: []types.Card{types.Creature}},
		{SubtypesAny: []types.Sub{types.Aura}},
	}}
	if issues := ValidateCardDef(spellTypeEffectCard(effect, zone.Battlefield)); len(issues) != 0 {
		t.Fatalf("AnyOf type-line selection issues = %+v, want none", issues)
	}
}

// TestValidateCardDefRejectsMalformedSpellTypeEffects proves a SpellSelection
// is only accepted on a battlefield static type-layer effect that selects by
// type line.
func TestValidateCardDefRejectsMalformedSpellTypeEffects(t *testing.T) {
	t.Parallel()
	nonTypeLayer := creatureSpellTypeEffect()
	nonTypeLayer.Layer = LayerColor
	nonTypeLayer.AddTypes = nil
	nonTypeLayer.AddColors = []color.Color{color.Blue}

	colorSelection := creatureSpellTypeEffect()
	colorSelection.SpellSelection = &Selection{ColorsAny: []color.Color{color.Red}}

	nestedColorSelection := creatureSpellTypeEffect()
	nestedColorSelection.SpellSelection = &Selection{AnyOf: []Selection{{ColorsAny: []color.Color{color.Red}}}}

	cardZone := creatureSpellTypeEffect()
	cardZone.CardZone = zone.Graveyard

	for name, test := range map[string]struct {
		effect ContinuousEffect
		zone   zone.Type
		issue  CardDefIssueCode
	}{
		"non-type layer":         {nonTypeLayer, zone.None, CardDefIssueInvalidAbilityBody},
		"color predicate":        {colorSelection, zone.None, CardDefIssueInvalidSelection},
		"nested color predicate": {nestedColorSelection, zone.None, CardDefIssueInvalidSelection},
		"card zone":              {cardZone, zone.None, CardDefIssueInvalidAbilityBody},
		"graveyard static":       {creatureSpellTypeEffect(), zone.Graveyard, CardDefIssueInvalidAbilityBody},
	} {
		if issues := ValidateCardDef(spellTypeEffectCard(test.effect, test.zone)); !hasCardDefIssue(issues, test.issue) {
			t.Errorf("%s: issues = %+v, want %s", name, issues, test.issue)
		}
	}
}

// TestApplyContinuousRejectsSpellTypeEffect proves spell type effects are not
// created by resolving instructions.
func TestApplyContinuousRejectsSpellTypeEffect(t *testing.T) {
	t.Parallel()
	primitive := ApplyContinuous{ContinuousEffects: []ContinuousEffect{creatureSpellTypeEffect()}}
	err := primitive.validatePrimitive(nil, false)
	if err == nil || !strings.Contains(err.Error(), "spell type effects") {
		t.Fatalf("validatePrimitive() = %v, want spell type effect rejection", err)
	}
}
