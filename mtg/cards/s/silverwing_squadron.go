package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SilverwingSquadron is the card definition for Silverwing Squadron.
//
// Type: Creature — Human Knight
// Cost: {5}{W}
//
// Oracle text:
//
//	Flying, vigilance
//	Silverwing Squadron's power and toughness are each equal to the number of creatures you control.
//	Whenever this creature attacks, create a number of 2/2 white Knight creature tokens with vigilance equal to the number of opponents you have.
var SilverwingSquadron = newSilverwingSquadron

func newSilverwingSquadron() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Silverwing Squadron",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:           []color.Color{color.White},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Human, types.Knight},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, create a number of 2/2 white Knight creature tokens with vigilance equal to the number of opponents you have.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountOpponentCount,
												Multiplier: 1,
											}),
											Source: game.TokenDef(silverwingSquadronToken),
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
			OracleText: `
			Flying, vigilance
			Silverwing Squadron's power and toughness are each equal to the number of creatures you control.
			Whenever this creature attacks, create a number of 2/2 white Knight creature tokens with vigilance equal to the number of opponents you have.
		`,
		},
	}
}

var silverwingSquadronToken = newSilverwingSquadronToken()

func newSilverwingSquadronToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Knight",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
		},
	}
}
