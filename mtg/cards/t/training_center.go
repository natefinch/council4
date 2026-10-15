package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TrainingCenter is the card definition for Training Center.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you have two or more opponents.
//	{T}: Add {U} or {R}.
var TrainingCenter = newTrainingCenter

func newTrainingCenter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:  "Training Center",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you have two or more opponents.", &game.Condition{
					Negate:     true,
					Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateOpponentCount, Op: compare.GreaterOrEqual, Value: 2}},
				}),
			},
			OracleText: `
			This land enters tapped unless you have two or more opponents.
			{T}: Add {U} or {R}.
		`,
		},
	}
}
