package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlimpseTheFuture is the card definition for Glimpse the Future.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	Look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard.
var GlimpseTheFuture = newGlimpseTheFuture

func newGlimpseTheFuture() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Glimpse the Future",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player: game.ControllerReference(),
							Look:   game.Fixed(3),
							Take:   game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard.
		`,
		},
	}
}
