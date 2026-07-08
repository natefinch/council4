package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GreaterTanuki is the card definition for Greater Tanuki.
//
// Type: Enchantment Creature — Dog
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Trample
//	Channel — {2}{G}, Discard this card: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
var GreaterTanuki = newGreaterTanuki

func newGreaterTanuki() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Greater Tanuki",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Dog},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "Channel — {2}{G}, Discard this card: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.G}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalDiscard,
							Text:   "Discard this card",
							Amount: 1,
							Source: zone.Hand,
						},
					},
					ZoneOfFunction: zone.Hand,
					Content: game.Mode{
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
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			Channel — {2}{G}, Discard this card: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.
		`,
		},
	}
}
