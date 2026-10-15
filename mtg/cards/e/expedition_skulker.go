package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ExpeditionSkulker is the card definition for Expedition Skulker.
//
// Type: Creature — Vampire Rogue
// Cost: {1}{B}
//
// Oracle text:
//
//	This creature has deathtouch as long as you control another Rogue. (Any amount of damage it deals to a creature is enough to destroy it.)
var ExpeditionSkulker = newExpeditionSkulker

func newExpeditionSkulker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Expedition Skulker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has deathtouch as long as you control another Rogue. (Any amount of damage it deals to a creature is enough to destroy it.)",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control another Rogue",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Rogue}, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Deathtouch},
						},
					},
				},
			},
			OracleText: `
			This creature has deathtouch as long as you control another Rogue. (Any amount of damage it deals to a creature is enough to destroy it.)
		`,
		},
	}
}
