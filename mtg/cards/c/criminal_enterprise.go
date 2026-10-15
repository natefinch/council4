package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CriminalEnterprise is the card definition for Criminal Enterprise.
//
// Type: Enchantment
// Cost: {2}{B}
//
// Oracle text:
//
//	When this enchantment enters, create a 2/1 black Villain creature token with menace. (It can't be blocked except by two or more creatures.)
//	Whenever a Villain you control dies, this enchantment deals 1 damage to each opponent and you gain 1 life.
var CriminalEnterprise = newCriminalEnterprise

func newCriminalEnterprise() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Criminal Enterprise",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this enchantment enters, create a 2/1 black Villain creature token with menace. (It can't be blocked except by two or more creatures.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
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
											Amount: game.Fixed(1),
											Source: game.TokenDef(criminalEnterpriseToken),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "Whenever a Villain you control dies, this enchantment deals 1 damage to each opponent and you gain 1 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Villain}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			When this enchantment enters, create a 2/1 black Villain creature token with menace. (It can't be blocked except by two or more creatures.)
			Whenever a Villain you control dies, this enchantment deals 1 damage to each opponent and you gain 1 life.
		`,
		},
	}
}

var criminalEnterpriseToken = newCriminalEnterpriseToken()

func newCriminalEnterpriseToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Villain",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Villain},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
		},
	}
}
