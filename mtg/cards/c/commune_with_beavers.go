package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CommuneWithBeavers is the card definition for Commune with Beavers.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Look at the top three cards of your library. You may reveal an artifact, creature, or land card from among them and put it into your hand. Put the rest on the bottom of your library in any order.
var CommuneWithBeavers = newCommuneWithBeavers()

func newCommuneWithBeavers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Commune with Beavers",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(3),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderLibraryBottom,
							Filter:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land}}),
							TakeUpTo:  true,
							Reveal:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Look at the top three cards of your library. You may reveal an artifact, creature, or land card from among them and put it into your hand. Put the rest on the bottom of your library in any order.
		`,
		},
	}
}
