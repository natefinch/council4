package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// LuxurySuite is the card definition for Luxury Suite.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you have two or more opponents.
//	{T}: Add {B} or {R}.
var LuxurySuite = newLuxurySuite

func newLuxurySuite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Luxury Suite",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you have two or more opponents.", &game.Condition{
					Negate:     true,
					Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateOpponentCount, Op: compare.GreaterOrEqual, Value: 2}},
				}),
			},
			OracleText: `
			This land enters tapped unless you have two or more opponents.
			{T}: Add {B} or {R}.
		`,
		},
	}
}
