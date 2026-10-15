package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DuneBroodNephilim is the card definition for Dune-Brood Nephilim.
//
// Type: Creature — Nephilim
// Cost: {B}{R}{G}{W}
//
// Oracle text:
//
//	Whenever this creature deals combat damage to a player, create a 1/1 colorless Sand creature token for each land you control.
var DuneBroodNephilim = newDuneBroodNephilim

func newDuneBroodNephilim() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Dune-Brood Nephilim",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Black, color.Green, color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Nephilim},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature deals combat damage to a player, create a 1/1 colorless Sand creature token for each land you control.",
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
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
											}),
											Source: game.TokenDef(duneBroodNephilimToken),
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
			Whenever this creature deals combat damage to a player, create a 1/1 colorless Sand creature token for each land you control.
		`,
		},
	}
}

var duneBroodNephilimToken = newDuneBroodNephilimToken()

func newDuneBroodNephilimToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Sand",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sand},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
