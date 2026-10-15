package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Windfall is the card definition for Windfall.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	Each player discards their hand, then draws cards equal to the greatest number of cards a player discarded this way.
var Windfall = newWindfall

func newWindfall() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Windfall",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							PlayerGroup: game.AllPlayersReference(),
							EntireHand:  true,
						},
						PublishResult: game.ResultKey("discarded-this-way"),
					},
					{
						Primitive: game.Draw{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:      game.DynamicAmountPreviousEffectResult,
								ResultKey: game.ResultKey("discarded-this-way"),
							}),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player discards their hand, then draws cards equal to the greatest number of cards a player discarded this way.
		`,
		},
	}
}
