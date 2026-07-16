package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TragicArrogance is the card definition for Tragic Arrogance.
//
// Type: Sorcery
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	For each player, you choose from among the permanents that player controls an artifact, a creature, an enchantment, and a planeswalker. Then each player sacrifices all other nonland permanents they control.
var TragicArrogance = newTragicArrogance

func newTragicArrogance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Tragic Arrogance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.KeepOnePerType{
							Players:                 game.AllPlayersReference(),
							Types:                   []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Planeswalker},
							AffectedSelection:       game.Selection{ExcludedTypes: []types.Card{types.Land}},
							ControllerChoosesForAll: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			For each player, you choose from among the permanents that player controls an artifact, a creature, an enchantment, and a planeswalker. Then each player sacrifices all other nonland permanents they control.
		`,
		},
	}
}
