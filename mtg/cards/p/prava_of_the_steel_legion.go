package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PravaOfTheSteelLegion is the card definition for Prava of the Steel Legion.
//
// Type: Legendary Creature — Cat Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	During your turn, creature tokens you control get +1/+4.
//	{3}{W}: Create a 1/1 white Soldier creature token.
//	Partner (You can have two commanders if both have partner.)
var PravaOfTheSteelLegion = newPravaOfTheSteelLegion

func newPravaOfTheSteelLegion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Prava of the Steel Legion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Cat, types.Soldier},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "During your turn, creature tokens you control get +1/+4.",
					Condition: opt.Val(game.Condition{
						Text:                 "During your turn",
						SourceCounterKind:    counter.PlusOnePlusOne,
						SourceControllerTurn: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 4,
						},
					},
				},
				game.PartnerStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{W}: Create a 1/1 white Soldier creature token.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.W,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(pravaOfTheSteelLegionToken),
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
			During your turn, creature tokens you control get +1/+4.
			{3}{W}: Create a 1/1 white Soldier creature token.
			Partner (You can have two commanders if both have partner.)
		`,
		},
	}
}

var pravaOfTheSteelLegionToken = newPravaOfTheSteelLegionToken()

func newPravaOfTheSteelLegionToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
