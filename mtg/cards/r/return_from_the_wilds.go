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

// ReturnFromTheWilds is the card definition for Return from the Wilds.
//
// Type: Sorcery
// Cost: {2}{G}
//
// Oracle text:
//
//	Choose two —
//	• Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
//	• Create a 1/1 white Human creature token.
//	• Create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
var ReturnFromTheWilds = newReturnFromTheWilds

func newReturnFromTheWilds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Return from the Wilds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:   zone.Library,
										Destination:  zone.Battlefield,
										Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
										EntersTapped: true,
									},
									Amount: game.Fixed(1),
								},
							},
						},
					},
					game.Mode{
						Text: "Create a 1/1 white Human creature token.",
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(returnFromTheWildsToken),
								},
							},
						},
					},
					game.Mode{
						Text: "Create a Food token. (It's an artifact with \"{2}, {T}, Sacrifice this token: You gain 3 life.\")",
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(returnFromTheWildsToken2),
								},
							},
						},
					},
				},
				MinModes: 2,
				MaxModes: 2,
			}),
			OracleText: `
			Choose two —
			• Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
			• Create a 1/1 white Human creature token.
			• Create a Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
		`,
		},
	}
}

var returnFromTheWildsToken = newReturnFromTheWildsToken()

func newReturnFromTheWildsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}

var returnFromTheWildsToken2 = newReturnFromTheWildsToken2()

func newReturnFromTheWildsToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Food",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Food},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(3),
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
		},
	}
}
