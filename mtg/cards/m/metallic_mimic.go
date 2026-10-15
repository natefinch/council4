package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MetallicMimic is the card definition for Metallic Mimic.
//
// Type: Artifact Creature — Shapeshifter
// Cost: {2}
//
// Oracle text:
//
//	As this creature enters, choose a creature type.
//	This creature is the chosen type in addition to its other types.
//	Each other creature you control of the chosen type enters with an additional +1/+1 counter on it.
var MetallicMimic = newMetallicMimic

func newMetallicMimic() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Metallic Mimic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature is the chosen type in addition to its other types.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:                  game.DurationPermanent,
							AffectedSource:            true,
							Layer:                     game.LayerType,
							AddSubtypeFromEntryChoice: game.EntryTypeChoiceKey,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this creature enters, choose a creature type."),
				game.EntersWithCountersGroupReplacement("Each other creature you control of the chosen type enters with an additional +1/+1 counter on it.", &game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry, Controller: game.ControllerYou, ExcludeSource: true}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			As this creature enters, choose a creature type.
			This creature is the chosen type in addition to its other types.
			Each other creature you control of the chosen type enters with an additional +1/+1 counter on it.
		`,
		},
	}
}
