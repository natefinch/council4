package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const theIndomitableOracleText = "Trample\n" +
	"Whenever a creature you control deals combat damage to a player, draw a card.\n" +
	"Crew 3\n" +
	"You may cast this card from your graveyard as long as you control three or more tapped Pirates and/or Vehicles."

func theIndomitableCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "The Indomitable",
		Layout:     "normal",
		ManaCost:   "{2}{U}{U}",
		TypeLine:   "Legendary Artifact — Vehicle",
		OracleText: theIndomitableOracleText,
		Power:      new("6"),
		Toughness:  new("6"),
	}
}

func TestLowerTheIndomitableGraveyardCastPermission(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, theIndomitableCard())

	var permission *game.StaticAbility
	for i := range face.StaticAbilities {
		body := &face.StaticAbilities[i].Body
		if body.ZoneOfFunction == zone.Graveyard {
			permission = body
			break
		}
	}
	if permission == nil {
		t.Fatalf("static abilities = %#v, want graveyard permission", face.StaticAbilities)
	}
	if !permission.Condition.Exists || !permission.Condition.Val.ControlsMatching.Exists {
		t.Fatalf("condition = %#v, want controls-matching gate", permission.Condition)
	}
	count := permission.Condition.Val.ControlsMatching.Val
	if count.MinCount != 3 ||
		count.Selection.Tapped != game.TriTrue ||
		!slices.Equal(count.Selection.SubtypesAny, []types.Sub{types.Pirate, types.Vehicle}) {
		t.Fatalf("condition count = %#v", count)
	}
	if len(permission.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", permission.RuleEffects)
	}
	effect := permission.RuleEffects[0]
	if effect.Kind != game.RuleEffectCastFromZone ||
		effect.AffectedPlayer != game.PlayerYou ||
		effect.CastFromZone != zone.Graveyard ||
		!effect.AffectedSource {
		t.Fatalf("rule effect = %#v", effect)
	}
}

func TestGenerateTheIndomitable(t *testing.T) {
	t.Parallel()
	generatedSourceContains(t, theIndomitableCard(), []string{
		"game.TrampleStaticBody",
		"game.CrewActivatedAbility(3)",
		"game.EventDamageDealt",
		"RequireCombatDamage:   true",
		"game.Draw{",
		"ZoneOfFunction: zone.Graveyard",
		"MinCount:  3",
		"Tapped: game.TriTrue",
		`types.Sub("Pirate")`,
		`types.Sub("Vehicle")`,
		"Kind:           game.RuleEffectCastFromZone",
		"CastFromZone:   zone.Graveyard",
		"AffectedSource: true",
	})
}
