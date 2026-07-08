package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GargoyleFlock is the card definition for Gargoyle Flock.
//
// Type: Creature — Tyranid Gargoyle
// Cost: {2}{G}{U}
//
// Oracle text:
//
//	Flying
//	Skyswarm — At the beginning of your end step, if a creature entered the battlefield under your control this turn, create a 1/1 blue Tyranid Gargoyle creature token with flying.
var GargoyleFlock = newGargoyleFlock

func newGargoyleFlock() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Gargoyle Flock",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Tyranid, types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningIf: "if a creature entered the battlefield under your control this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:            game.EventPermanentEnteredBattlefield,
								Controller:       game.TriggerControllerYou,
								SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(gargoyleFlockToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Skyswarm — At the beginning of your end step, if a creature entered the battlefield under your control this turn, create a 1/1 blue Tyranid Gargoyle creature token with flying.
		`,
		},
	}
}

var gargoyleFlockToken = newGargoyleFlockToken()

func newGargoyleFlockToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Tyranid Gargoyle",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Tyranid, types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
