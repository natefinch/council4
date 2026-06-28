package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LairDelve is the card definition for Lair Delve.
//
// Type: Sorcery
// Cost: {2}{G}
//
// Oracle text:
//
//	Reveal the top two cards of your library. Put all creature and land cards revealed this way into your hand and the rest on the bottom of your library in any order.
var LairDelve = newLairDelve()

func newLairDelve() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lair Delve",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.RevealTopPartition{
							Player:    game.ControllerReference(),
							Amount:    game.Fixed(2),
							Selection: game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}},
							Remainder: game.DigRemainderLibraryBottom,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Reveal the top two cards of your library. Put all creature and land cards revealed this way into your hand and the rest on the bottom of your library in any order.
		`,
		},
	}
}
