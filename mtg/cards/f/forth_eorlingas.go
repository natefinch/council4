package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ForthEorlingas is the card definition for Forth Eorlingas!.
//
// Type: Sorcery
// Cost: {X}{R}{W}
//
// Oracle text:
//
//	Create X 2/2 red Human Knight creature tokens with trample and haste.
//	Whenever one or more creatures you control deal combat damage to one or more players this turn, you become the monarch.
var ForthEorlingas = newForthEorlingas

func newForthEorlingas() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Forth Eorlingas!",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Source: game.TokenDef(forthEorlingasToken),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								EventPattern: opt.Val(game.TriggerPattern{
									Event:                 game.EventDamageDealt,
									Controller:            game.TriggerControllerYou,
									Subject:               game.TriggerSubjectDamageSource,
									OneOrMore:             true,
									RequireCombatDamage:   true,
									DamageRecipient:       game.DamageRecipientPlayer,
									DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								}),
								Window: game.DelayedWindowThisTurn,
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.BecomeMonarch{
												Player: game.ControllerReference(),
											},
										},
									},
								}.Ability(),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create X 2/2 red Human Knight creature tokens with trample and haste.
			Whenever one or more creatures you control deal combat damage to one or more players this turn, you become the monarch.
		`,
		},
	}
}

var forthEorlingasToken = newForthEorlingasToken()

func newForthEorlingasToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human Knight",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.HasteStaticBody,
			},
		},
	}
}
