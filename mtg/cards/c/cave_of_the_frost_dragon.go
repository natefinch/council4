package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CaveOfTheFrostDragon is the card definition for Cave of the Frost Dragon.
//
// Type: Land
//
// Oracle text:
//
//	If you control two or more other lands, this land enters tapped.
//	{T}: Add {W}.
//	{4}{W}: This land becomes a 3/4 white Dragon creature with flying until end of turn. It's still a land.
var CaveOfTheFrostDragon = newCaveOfTheFrostDragon

func newCaveOfTheFrostDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:  "Cave of the Frost Dragon",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{4}{W}: This land becomes a 3/4 white Dragon creature with flying until end of turn. It's still a land.",
					ManaCost:       opt.Val(cost.Mana{cost.O(4), cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourcePermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:     game.LayerColor,
											SetColors: []color.Color{color.White},
										},
										game.ContinuousEffect{
											Layer:       game.LayerType,
											AddTypes:    []types.Card{types.Creature},
											AddSubtypes: []types.Sub{types.Dragon},
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Flying,
											},
										},
										game.ContinuousEffect{
											Layer:        game.LayerPowerToughnessSet,
											SetPower:     opt.Val(game.PT{Value: 3}),
											SetToughness: opt.Val(game.PT{Value: 4}),
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("If you control two or more other lands, this land enters tapped.", &game.Condition{
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, ExcludeSource: true},
						MinCount:  2,
					}),
				}),
			},
			OracleText: `
			If you control two or more other lands, this land enters tapped.
			{T}: Add {W}.
			{4}{W}: This land becomes a 3/4 white Dragon creature with flying until end of turn. It's still a land.
		`,
		},
	}
}
