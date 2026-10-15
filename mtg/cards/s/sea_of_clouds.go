package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SeaOfClouds is the card definition for Sea of Clouds.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you have two or more opponents.
//	{T}: Add {W} or {U}.
var SeaOfClouds = newSeaOfClouds

func newSeaOfClouds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:  "Sea of Clouds",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you have two or more opponents.", &game.Condition{
					Negate:     true,
					Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateOpponentCount, Op: compare.GreaterOrEqual, Value: 2}},
				}),
			},
			OracleText: `
			This land enters tapped unless you have two or more opponents.
			{T}: Add {W} or {U}.
		`,
		},
	}
}
