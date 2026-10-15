package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DayOfJudgment is the card definition for Day of Judgment.
//
// Type: Sorcery
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Destroy all creatures.
var DayOfJudgment = newDayOfJudgment

func newDayOfJudgment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Day of Judgment",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all creatures.
		`,
		},
	}
}
