package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCloudKeyChosenCardTypeReduction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCostModifier,
			AffectedPlayer: game.PlayerYou,
			CostModifier: game.CostModifier{
				Kind:                          game.CostModifierSpell,
				GenericReduction:              1,
				ChosenCardTypeFromEntryChoice: true,
			},
		}}}},
	}})
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryCardTypeChoiceKey: {
			Kind:     game.ResolutionChoiceCardType,
			CardType: types.Artifact,
		},
	}
	artifactCreature := &game.CardDef{CardFace: game.CardFace{
		Types: []types.Card{types.Artifact, types.Creature},
	}}
	creature := &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}}
	if got := len(staticCostModifiersForContext(g, game.Player1, artifactCreature, zone.Hand, nil)); got != 1 {
		t.Fatalf("artifact modifiers = %d, want 1", got)
	}
	if got := len(staticCostModifiersForContext(g, game.Player1, creature, zone.Hand, nil)); got != 0 {
		t.Fatalf("creature modifiers = %d, want 0", got)
	}
}
