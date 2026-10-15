package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EverbarkShaman is the card definition for Everbark Shaman.
//
// Type: Creature — Treefolk Shaman
// Cost: {4}{G}
//
// Oracle text:
//
//	{T}, Exile a Treefolk card from your graveyard: Search your library for up to two Forest cards, put them onto the battlefield tapped, then shuffle.
var EverbarkShaman = newEverbarkShaman

func newEverbarkShaman() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Everbark Shaman",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk, types.Shaman},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Exile a Treefolk card from your graveyard: Search your library for up to two Forest cards, put them onto the battlefield tapped, then shuffle.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalExile,
							Text:        "Exile a Treefolk card from your graveyard",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{0: types.Treefolk},
							Source:      zone.Graveyard,
							CounterKind: counter.PlusOnePlusOne,
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
												SourceZone:   zone.Library,
												Destination:  zone.Battlefield,
												Filter:       game.Selection{SubtypesAny: []types.Sub{types.Forest}},
												EntersTapped: true,
											},
											Amount: game.Fixed(2),
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
			{T}, Exile a Treefolk card from your graveyard: Search your library for up to two Forest cards, put them onto the battlefield tapped, then shuffle.
		`,
		},
	}
}
