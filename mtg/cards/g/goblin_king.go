package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinKing is the card definition for Goblin King.
//
// Type: Creature — Goblin
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Other Goblins get +1/+1 and have mountainwalk.
var GoblinKing = newGoblinKing

func newGoblinKing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin King",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Other Goblins get +1/+1 and have mountainwalk.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.BattlefieldGroupExcluding(game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin")}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.BattlefieldGroupExcluding(game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin")}}, game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.StaticAbility{
									Text: "Mountainwalk",
									KeywordAbilities: []game.KeywordAbility{
										game.LandwalkKeyword{
											Subtype: types.Mountain,
										},
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			Other Goblins get +1/+1 and have mountainwalk.
		`,
		},
	}
}
