package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RatadrabikOfUrborg is the card definition for Ratadrabik of Urborg.
//
// Type: Legendary Creature — Zombie Wizard
// Cost: {2}{W}{B}
//
// Oracle text:
//
//	Vigilance, ward {2}
//	Other Zombies you control have vigilance.
//	Whenever another legendary creature you control dies, create a token that's a copy of that creature, except it's not legendary and it's a 2/2 black Zombie in addition to its other colors and types.
var RatadrabikOfUrborg = newRatadrabikOfUrborg

func newRatadrabikOfUrborg() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Ratadrabik of Urborg",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Zombie, types.Wizard},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Zombie")}}, game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Vigilance,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source:          game.TokenCopySourceObject,
										Object:          game.EventPermanentReference(),
										SetNotLegendary: true,
										SetPower:        opt.Val(game.PT{Value: 2}),
										SetToughness:    opt.Val(game.PT{Value: 2}),
										AddColors:       []color.Color{color.Black},
										AddSubtypes:     []types.Sub{types.Zombie},
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance, ward {2}
			Other Zombies you control have vigilance.
			Whenever another legendary creature you control dies, create a token that's a copy of that creature, except it's not legendary and it's a 2/2 black Zombie in addition to its other colors and types.
		`,
		},
	}
}
