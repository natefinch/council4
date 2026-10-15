package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TivadarSCrusade is the card definition for Tivadar's Crusade.
//
// Type: Sorcery
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	Destroy all Goblins.
var TivadarSCrusade = newTivadarSCrusade

func newTivadarSCrusade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Tivadar's Crusade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin")}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all Goblins.
		`,
		},
	}
}
