package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WaveGoodbye is the card definition for Wave Goodbye.
//
// Type: Sorcery
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Return each creature without a +1/+1 counter on it to its owner's hand.
var WaveGoodbye = newWaveGoodbye

func newWaveGoodbye() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Wave Goodbye",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchExcludedCounter: true, ExcludedCounter: counter.PlusOnePlusOne}),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return each creature without a +1/+1 counter on it to its owner's hand.
		`,
		},
	}
}
