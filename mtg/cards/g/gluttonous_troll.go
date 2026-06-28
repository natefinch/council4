package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GluttonousTroll is the card definition for Gluttonous Troll.
//
// Type: Creature — Troll
//
// Oracle text:
//
//	Trample
//	When this creature enters, create a number of Food tokens equal to the number of opponents you have. (Food tokens are artifacts with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	{1}{G}, Sacrifice another nonland permanent: This creature gets +2/+2 until end of turn.
var GluttonousTroll = newGluttonousTroll()

func newGluttonousTroll() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Gluttonous Troll",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Troll},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{G}, Sacrifice another nonland permanent: This creature gets +2/+2 until end of turn.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:                 cost.AdditionalSacrifice,
							Text:                 "Sacrifice another nonland permanent",
							Amount:               1,
							ExcludePermanentType: types.Land,
							ExcludeSource:        true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(2),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountOpponentCount,
										Multiplier: 1,
									}),
									Source: game.TokenDef(gluttonousTrollToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			When this creature enters, create a number of Food tokens equal to the number of opponents you have. (Food tokens are artifacts with "{2}, {T}, Sacrifice this token: You gain 3 life.")
			{1}{G}, Sacrifice another nonland permanent: This creature gets +2/+2 until end of turn.
		`,
		},
	}
}

var gluttonousTrollToken = newGluttonousTrollToken()

func newGluttonousTrollToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Food",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Food},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
