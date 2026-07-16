package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cataclysm is the card definition for Cataclysm.
//
// Type: Sorcery
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Each player chooses from among the permanents they control an artifact, a creature, an enchantment, and a land, then sacrifices the rest.
var Cataclysm = newCataclysm

func newCataclysm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Cataclysm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.KeepOnePerType{
							Players: game.AllPlayersReference(),
							Types:   []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player chooses from among the permanents they control an artifact, a creature, an enchantment, and a land, then sacrifices the rest.
		`,
		},
	}
}
