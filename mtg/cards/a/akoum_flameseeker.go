package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AkoumFlameseeker is the card definition for Akoum Flameseeker.
//
// Type: Creature — Human Shaman Ally
// Cost: {2}{R}
//
// Oracle text:
//
//	Cohort — {T}, Tap an untapped Ally you control: Discard a card. If you do, draw a card.
var AkoumFlameseeker = newAkoumFlameseeker

func newAkoumFlameseeker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Akoum Flameseeker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Shaman, types.Ally},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Cohort — {T}, Tap an untapped Ally you control: Discard a card. If you do, draw a card.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap an untapped Ally you control",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Ally},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Cohort — {T}, Tap an untapped Ally you control: Discard a card. If you do, draw a card.
		`,
		},
	}
}
