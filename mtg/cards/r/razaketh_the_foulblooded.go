package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RazakethTheFoulblooded is the card definition for Razaketh, the Foulblooded.
//
// Type: Legendary Creature — Demon
// Cost: {5}{B}{B}{B}
//
// Oracle text:
//
//	Flying, trample
//	Pay 2 life, Sacrifice another creature: Search your library for a card, put that card into your hand, then shuffle.
var RazakethTheFoulblooded = newRazakethTheFoulblooded

func newRazakethTheFoulblooded() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Razaketh, the Foulblooded",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Demon},
			Power:      opt.Val(game.PT{Value: 8}),
			Toughness:  opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Pay 2 life, Sacrifice another creature: Search your library for a card, put that card into your hand, then shuffle.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalPayLife,
							Text:        "Pay 2 life",
							Amount:      2,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice another creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:       zone.Library,
												Destination:      zone.Hand,
												FailToFindPolicy: game.SearchMustFindIfAvailable,
											},
											Amount: game.Fixed(1),
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
			Flying, trample
			Pay 2 life, Sacrifice another creature: Search your library for a card, put that card into your hand, then shuffle.
		`,
		},
	}
}
