package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BurningInquiry is the card definition for Burning Inquiry.
//
// Type: Sorcery
// Cost: {R}
//
// Oracle text:
//
//	Each player draws three cards, then discards three cards at random.
var BurningInquiry = newBurningInquiry

func newBurningInquiry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Burning Inquiry",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount:      game.Fixed(3),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
					{
						Primitive: game.Discard{
							Amount:      game.Fixed(3),
							PlayerGroup: game.AllPlayersReference(),
							AtRandom:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player draws three cards, then discards three cards at random.
		`,
		},
	}
}
