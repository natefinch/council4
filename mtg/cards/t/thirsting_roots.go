package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ThirstingRoots is the card definition for Thirsting Roots.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Choose one —
//	• Search your library for a basic land card, reveal it, put it into your hand, then shuffle.
//	• Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
var ThirstingRoots = newThirstingRoots

func newThirstingRoots() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Thirsting Roots",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Hand,
										Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
										Reveal:      true,
									},
									Amount: game.Fixed(1),
								},
							},
						},
					},
					game.Mode{
						Text: "Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)",
						Sequence: []game.Instruction{
							{
								Primitive: game.Proliferate{
									Amount: game.Fixed(1),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Search your library for a basic land card, reveal it, put it into your hand, then shuffle.
			• Proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)
		`,
		},
	}
}
