package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheWorldTree is the card definition for The World Tree.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {G}.
//	As long as you control six or more lands, lands you control have "{T}: Add one mana of any color."
//	{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, {T}, Sacrifice this land: Search your library for any number of God cards, put them onto the battlefield, then shuffle.
var TheWorldTree = newTheWorldTree

func newTheWorldTree() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name:  "The World Tree",
			Types: []types.Card{types.Land},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
							MinCount:  6,
						}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Land}}),
							AddAbilities: []game.Ability{
								new(game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G)),
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, {T}, Sacrifice this land: Search your library for any number of God cards, put them onto the battlefield, then shuffle.",
					ManaCost: opt.Val(cost.Mana{cost.W, cost.W, cost.U, cost.U, cost.B, cost.B, cost.R, cost.R, cost.G, cost.G}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this land",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:  zone.Library,
										Destination: zone.Battlefield,
										Filter:      game.Selection{SubtypesAny: []types.Sub{types.Sub("God")}},
										AnyNumber:   true,
									},
									Amount: game.Fixed(0),
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {G}.
			As long as you control six or more lands, lands you control have "{T}: Add one mana of any color."
			{W}{W}{U}{U}{B}{B}{R}{R}{G}{G}, {T}, Sacrifice this land: Search your library for any number of God cards, put them onto the battlefield, then shuffle.
		`,
		},
	}
}
