package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// UndergrowthStadium is the card definition for Undergrowth Stadium.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you have two or more opponents.
//	{T}: Add {B} or {G}.
var UndergrowthStadium = newUndergrowthStadium

func newUndergrowthStadium() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:  "Undergrowth Stadium",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you have two or more opponents.", &game.Condition{
					Negate:     true,
					Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateOpponentCount, Op: compare.GreaterOrEqual, Value: 2}},
				}),
			},
			OracleText: `
			This land enters tapped unless you have two or more opponents.
			{T}: Add {B} or {G}.
		`,
		},
	}
}
