package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DromokaSGift is the card definition for Dromoka's Gift.
//
// Type: Instant
// Cost: {4}{G}
//
// Oracle text:
//
//	Bolster 4. (Choose a creature with the least toughness among creatures you control and put four +1/+1 counters on it.)
var DromokaSGift = newDromokaSGift

func newDromokaSGift() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Dromoka's Gift",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Bolster{
							Amount: game.Fixed(4),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Bolster 4. (Choose a creature with the least toughness among creatures you control and put four +1/+1 counters on it.)
		`,
		},
	}
}
