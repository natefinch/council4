package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DanceOfTheTumbleweeds is the card definition for Dance of the Tumbleweeds.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Spree (Choose one or more additional costs.)
//	+ {1} — Search your library for a basic land card or a Desert card, put it onto the battlefield, then shuffle.
//	+ {3} — Create an X/X green Elemental creature token, where X is the number of lands you control.
var DanceOfTheTumbleweeds = newDanceOfTheTumbleweeds()

func newDanceOfTheTumbleweeds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Dance of the Tumbleweeds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "{1} — Search your library for a basic land card or a Desert card, put it onto the battlefield, then shuffle.",
						Cost: opt.Val(cost.Mana{cost.O(1)}),
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Battlefield,
										Filter:      game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Desert")}}}},
									},
									Amount: game.Fixed(1),
								},
							},
						},
					},
					game.Mode{
						Text: "{3} — Create an X/X green Elemental creature token, where X is the number of lands you control.",
						Cost: opt.Val(cost.Mana{cost.O(3)}),
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(danceOfTheTumbleweedsToken),
									Power: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
									})),
									Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
									})),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 2,
			}),
			OracleText: `
			Spree (Choose one or more additional costs.)
			+ {1} — Search your library for a basic land card or a Desert card, put it onto the battlefield, then shuffle.
			+ {3} — Create an X/X green Elemental creature token, where X is the number of lands you control.
		`,
		},
	}
}

var danceOfTheTumbleweedsToken = newDanceOfTheTumbleweedsToken()

func newDanceOfTheTumbleweedsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Elemental",
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Elemental},
		},
	}
}
