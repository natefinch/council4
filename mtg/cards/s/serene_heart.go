package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SereneHeart is the card definition for Serene Heart.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Destroy all Auras.
var SereneHeart = newSereneHeart

func newSereneHeart() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Serene Heart",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Aura")}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all Auras.
		`,
		},
	}
}
