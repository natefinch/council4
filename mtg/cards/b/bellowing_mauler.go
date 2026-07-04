package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BellowingMauler is the card definition for Bellowing Mauler.
//
// Type: Creature — Ogre Warrior
// Cost: {4}{B}
//
// Oracle text:
//
//	At the beginning of your end step, each player loses 4 life unless they sacrifice a nontoken creature of their choice.
var BellowingMauler = newBellowingMauler()

func newBellowingMauler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bellowing Mauler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ogre, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PunisherEachLoseLife{
									PlayerGroup:        game.AllPlayersReference(),
									Amount:             game.Fixed(4),
									AllowSacrifice:     true,
									SacrificeSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your end step, each player loses 4 life unless they sacrifice a nontoken creature of their choice.
		`,
		},
	}
}
