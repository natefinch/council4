package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RoostOfDrakes is the card definition for Roost of Drakes.
//
// Type: Enchantment
// Cost: {U}
//
// Oracle text:
//
//	Kicker {2}{U} (You may pay an additional {2}{U} as you cast this spell.)
//	When this enchantment enters, if it was kicked, create a 2/2 blue Drake creature token with flying.
//	Whenever you cast a kicked spell, create a 2/2 blue Drake creature token with flying.
var RoostOfDrakes = newRoostOfDrakes

func newRoostOfDrakes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Roost of Drakes",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Kicker {2}{U}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.U,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this enchantment enters, if it was kicked, create a 2/2 blue Drake creature token with flying.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf:                        "if it was kicked",
						InterveningIfEventPermanentWasKicked: true,
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(roostOfDrakesToken),
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
					Text: "Whenever you cast a kicked spell, create a 2/2 blue Drake creature token with flying.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:             game.EventSpellCast,
							Controller:        game.TriggerControllerYou,
							StackObjectKind:   game.StackSpell,
							RequireKickerPaid: true,
							CounterKind:       counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(roostOfDrakesToken),
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
			Kicker {2}{U} (You may pay an additional {2}{U} as you cast this spell.)
			When this enchantment enters, if it was kicked, create a 2/2 blue Drake creature token with flying.
			Whenever you cast a kicked spell, create a 2/2 blue Drake creature token with flying.
		`,
		},
	}
}

var roostOfDrakesToken = newRoostOfDrakesToken()

func newRoostOfDrakesToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Drake",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Drake},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
