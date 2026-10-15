package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmpyrialStorm is the card definition for Empyrial Storm.
//
// Type: Sorcery
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	When you cast this spell, copy it for each time you've cast your commander from the command zone this game.
//	Create a 4/4 white Angel creature token with flying.
var EmpyrialStorm = newEmpyrialStorm

func newEmpyrialStorm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Empyrial Storm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When you cast this spell, copy it for each time you've cast your commander from the command zone this game.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							Source:          game.TriggerSourceSelf,
							SelfWasCast:     true,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CopyStackObject{
											Object: game.EventStackObjectReference(),
											DynamicCount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCommanderCastCount,
												Multiplier: 1,
											}),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(empyrialStormToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			When you cast this spell, copy it for each time you've cast your commander from the command zone this game.
			Create a 4/4 white Angel creature token with flying.
		`,
		},
	}
}

var empyrialStormToken = newEmpyrialStormToken()

func newEmpyrialStormToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Angel",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
