package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MnemonicNexus is the card definition for Mnemonic Nexus.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Each player shuffles their graveyard into their library.
var MnemonicNexus = newMnemonicNexus

func newMnemonicNexus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mnemonic Nexus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ShuffleGraveyardIntoLibrary{
							PlayerGroup: game.AllPlayersReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player shuffles their graveyard into their library.
		`,
		},
	}
}
