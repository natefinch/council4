package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RooftopBypass is the card definition for Rooftop Bypass.
//
// Type: Enchantment
// Cost: {1}{U}{B}
//
// Oracle text:
//
//	Whenever one or more nontoken creatures you control deal combat damage to a player, create a 1/1 black Assassin creature token with menace. (It can't be blocked except by two or more creatures.)
var RooftopBypass = newRooftopBypass

func newRooftopBypass() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Rooftop Bypass",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever one or more nontoken creatures you control deal combat damage to a player, create a 1/1 black Assassin creature token with menace. (It can't be blocked except by two or more creatures.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							StackObjectKind:       game.StackSpell,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
							RequireCombatDamage:   true,
							OneOrMore:             true,
							CounterKind:           counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(rooftopBypassToken),
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
			Whenever one or more nontoken creatures you control deal combat damage to a player, create a 1/1 black Assassin creature token with menace. (It can't be blocked except by two or more creatures.)
		`,
		},
	}
}

var rooftopBypassToken = newRooftopBypassToken()

func newRooftopBypassToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Assassin",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Assassin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
		},
	}
}
