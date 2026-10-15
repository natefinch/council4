package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuartzwoodCrasher is the card definition for Quartzwood Crasher.
//
// Type: Creature — Dinosaur Beast
// Cost: {2}{R}{R}{G}
//
// Oracle text:
//
//	Trample
//	Whenever one or more creatures you control with trample deal combat damage to a player, create an X/X green Dinosaur Beast creature token with trample, where X is the amount of damage those creatures dealt to that player.
var QuartzwoodCrasher = newQuartzwoodCrasher

func newQuartzwoodCrasher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Quartzwood Crasher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur, types.Beast},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever one or more creatures you control with trample deal combat damage to a player, create an X/X green Dinosaur Beast creature token with trample, where X is the amount of damage those creatures dealt to that player.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                     game.EventDamageDealt,
							Controller:                game.TriggerControllerYou,
							Subject:                   game.TriggerSubjectDamageSource,
							StackObjectKind:           game.StackSpell,
							DamageRecipient:           game.DamageRecipientPlayer,
							DamageSourceSelection:     game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Trample},
							RequireCombatDamage:       true,
							OneOrMore:                 true,
							OneOrMorePerDamagedPlayer: true,
							CounterKind:               counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(quartzwoodCrasherToken),
											Power: opt.Val(game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountTriggeringEventTotalCombatDamage,
												Multiplier: 1,
											})),
											Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountTriggeringEventTotalCombatDamage,
												Multiplier: 1,
											})),
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
			Whenever one or more creatures you control with trample deal combat damage to a player, create an X/X green Dinosaur Beast creature token with trample, where X is the amount of damage those creatures dealt to that player.
		`,
		},
	}
}

var quartzwoodCrasherToken = newQuartzwoodCrasherToken()

func newQuartzwoodCrasherToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Dinosaur Beast",
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Dinosaur, types.Beast},
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
		},
	}
}
