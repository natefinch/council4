package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DrillSkimmer is the card definition for Drill-Skimmer.
//
// Type: Artifact Creature — Thopter
// Cost: {4}
//
// Oracle text:
//
//	Flying
//	This creature has shroud as long as you control another artifact creature. (It can't be the target of spells or abilities.)
var DrillSkimmer = newDrillSkimmer

func newDrillSkimmer() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Drill-Skimmer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Thopter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "This creature has shroud as long as you control another artifact creature. (It can't be the target of spells or abilities.)",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control another artifact creature",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Shroud},
						},
					},
				},
			},
			OracleText: `
			Flying
			This creature has shroud as long as you control another artifact creature. (It can't be the target of spells or abilities.)
		`,
		},
	}
}
