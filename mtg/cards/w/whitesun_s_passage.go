package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WhitesunSPassage is the card definition for Whitesun's Passage.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	You gain 5 life.
var WhitesunSPassage = newWhitesunSPassage

func newWhitesunSPassage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Whitesun's Passage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(5),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You gain 5 life.
		`,
		},
	}
}
