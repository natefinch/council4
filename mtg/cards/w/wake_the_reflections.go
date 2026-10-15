package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WakeTheReflections is the card definition for Wake the Reflections.
//
// Type: Sorcery
// Cost: {W}
//
// Oracle text:
//
//	Populate. (Create a token that's a copy of a creature token you control.)
var WakeTheReflections = newWakeTheReflections

func newWakeTheReflections() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Wake the Reflections",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source: game.TokenCopySourceChosenControlledCreatureToken,
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Populate. (Create a token that's a copy of a creature token you control.)
		`,
		},
	}
}
