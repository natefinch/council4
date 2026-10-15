package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RadagastWizardOfWilds is the card definition for Radagast, Wizard of Wilds.
//
// Type: Legendary Creature — Avatar Wizard
// Cost: {2}{G}{U}
//
// Oracle text:
//
//	Ward {1}
//	Beasts and Birds you control have ward {1}.
//	Whenever you cast a spell with mana value 5 or greater, choose one —
//	• Create a 3/3 green Beast creature token.
//	• Create a 2/2 blue Bird creature token with flying.
var RadagastWizardOfWilds = newRadagastWizardOfWilds

func newRadagastWizardOfWilds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Radagast, Wizard of Wilds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Avatar, types.Wizard},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbility(cost.Mana{cost.O(1)}),
				game.StaticAbility{
					Text: "Beasts and Birds you control have ward {1}.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Beast"), types.Sub("Bird")}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.WardStaticAbility(cost.Mana{cost.O(1)})),
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a spell with mana value 5 or greater, choose one —\n• Create a 3/3 green Beast creature token.\n• Create a 2/2 blue Bird creature token with flying.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventSpellCast,
							Controller: game.TriggerControllerYou,
							CardSelection: game.Selection{ManaValue: opt.Val(compare.Int{
								Op:    compare.GreaterOrEqual,
								Value: 5,
							})},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Create a 3/3 green Beast creature token.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(radagastWizardOfWildsToken),
										},
									},
								},
							},
							game.Mode{
								Text: "Create a 2/2 blue Bird creature token with flying.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(radagastWizardOfWildsToken2),
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
			Ward {1}
			Beasts and Birds you control have ward {1}.
			Whenever you cast a spell with mana value 5 or greater, choose one —
			• Create a 3/3 green Beast creature token.
			• Create a 2/2 blue Bird creature token with flying.
		`,
		},
	}
}

var radagastWizardOfWildsToken = newRadagastWizardOfWildsToken()

func newRadagastWizardOfWildsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}

var radagastWizardOfWildsToken2 = newRadagastWizardOfWildsToken2()

func newRadagastWizardOfWildsToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bird",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
