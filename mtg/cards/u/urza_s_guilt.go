package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UrzaSGuilt is the card definition for Urza's Guilt.
//
// Type: Sorcery
// Cost: {2}{U}{B}
//
// Oracle text:
//
//	Each player draws two cards, then discards three cards, then loses 4 life.
var UrzaSGuilt = newUrzaSGuilt

func newUrzaSGuilt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Urza's Guilt",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount:      game.Fixed(2),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
					{
						Primitive: game.Discard{
							Amount:      game.Fixed(3),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount:      game.Fixed(4),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player draws two cards, then discards three cards, then loses 4 life.
		`,
		},
	}
}
