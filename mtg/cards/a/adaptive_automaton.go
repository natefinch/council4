package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AdaptiveAutomaton is the card definition for Adaptive Automaton.
//
// Type: Artifact Creature — Construct
// Cost: {3}
//
// Oracle text:
//
//	As this creature enters, choose a creature type.
//	This creature is the chosen type in addition to its other types.
//	Other creatures you control of the chosen type get +1/+1.
var AdaptiveAutomaton = newAdaptiveAutomaton

func newAdaptiveAutomaton() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Adaptive Automaton",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
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
				game.StaticAbility{
					Text: "Other creatures you control of the chosen type get +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this creature enters, choose a creature type."),
			},
			OracleText: `
			As this creature enters, choose a creature type.
			This creature is the chosen type in addition to its other types.
			Other creatures you control of the chosen type get +1/+1.
		`,
		},
	}
}
