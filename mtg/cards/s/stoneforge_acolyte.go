package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// StoneforgeAcolyte is the card definition for Stoneforge Acolyte.
//
// Type: Creature — Kor Artificer Ally
// Cost: {W}
//
// Oracle text:
//
//	Cohort — {T}, Tap an untapped Ally you control: Look at the top four cards of your library. You may reveal an Equipment card from among them and put it into your hand. Put the rest on the bottom of your library in any order.
var StoneforgeAcolyte = newStoneforgeAcolyte

func newStoneforgeAcolyte() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Stoneforge Acolyte",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kor, types.Artificer, types.Ally},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Cohort — {T}, Tap an untapped Ally you control: Look at the top four cards of your library. You may reveal an Equipment card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap an untapped Ally you control",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{0: types.Ally},
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Dig{
											Player:    game.ControllerReference(),
											Look:      game.Fixed(4),
											Take:      game.Fixed(1),
											Remainder: game.DigRemainderLibraryBottom,
											Filter:    opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Equipment}}),
											TakeUpTo:  true,
											Reveal:    true,
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
			Cohort — {T}, Tap an untapped Ally you control: Look at the top four cards of your library. You may reveal an Equipment card from among them and put it into your hand. Put the rest on the bottom of your library in any order.
		`,
		},
	}
}
