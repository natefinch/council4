package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThranGolem is the card definition for Thran Golem.
//
// Type: Artifact Creature — Golem
// Cost: {5}
//
// Oracle text:
//
//	As long as this creature is enchanted, it gets +2/+2 and has flying, first strike, and trample.
var ThranGolem = newThranGolem()

func newThranGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Thran Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchEnchanted: true}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddKeywords: []game.Keyword{
								game.Flying,
								game.FirstStrike,
								game.Trample,
							},
						},
					},
				},
			},
			OracleText: `
			As long as this creature is enchanted, it gets +2/+2 and has flying, first strike, and trample.
		`,
		},
	}
}
