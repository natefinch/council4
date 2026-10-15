package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BarterInBlood is the card definition for Barter in Blood.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Each player sacrifices two creatures of their choice.
var BarterInBlood = newBarterInBlood

func newBarterInBlood() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Barter in Blood",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							PlayerGroup: game.AllPlayersReference(),
							Amount:      game.Fixed(2),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player sacrifices two creatures of their choice.
		`,
		},
	}
}
