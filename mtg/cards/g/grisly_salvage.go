package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GrislySalvage is the card definition for Grisly Salvage.
//
// Type: Instant
// Cost: {B}{G}
//
// Oracle text:
//
//	Reveal the top five cards of your library. You may put a creature or land card from among them into your hand. Put the rest into your graveyard.
var GrislySalvage = newGrislySalvage

func newGrislySalvage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Grisly Salvage",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:   game.ControllerReference(),
							Look:     game.Fixed(5),
							Take:     game.Fixed(1),
							Filter:   opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}}),
							TakeUpTo: true,
							Reveal:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Reveal the top five cards of your library. You may put a creature or land card from among them into your hand. Put the rest into your graveyard.
		`,
		},
	}
}
