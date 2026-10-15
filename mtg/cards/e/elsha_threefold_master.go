package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElshaThreefoldMaster is the card definition for Elsha, Threefold Master.
//
// Type: Legendary Creature — Djinn Monk
// Cost: {U}{R}{W}
//
// Oracle text:
//
//	Trample
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	Whenever Elsha deals combat damage to a player, create that many 1/1 white Monk creature tokens with prowess.
var ElshaThreefoldMaster = newElshaThreefoldMaster

func newElshaThreefoldMaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Elsha, Threefold Master",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Djinn, types.Monk},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.ProwessStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Elsha deals combat damage to a player, create that many 1/1 white Monk creature tokens with prowess.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							StackObjectKind:     game.StackSpell,
							DamageRecipient:     game.DamageRecipientPlayer,
							RequireCombatDamage: true,
							CounterKind:         counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountEventDamage,
												Multiplier: 1,
											}),
											Source: game.TokenDef(elshaThreefoldMasterToken),
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
			Trample
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			Whenever Elsha deals combat damage to a player, create that many 1/1 white Monk creature tokens with prowess.
		`,
		},
	}
}

var elshaThreefoldMasterToken = newElshaThreefoldMasterToken()

func newElshaThreefoldMasterToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Monk",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Monk},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProwessStaticBody,
			},
		},
	}
}
