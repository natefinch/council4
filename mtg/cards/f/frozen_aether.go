package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FrozenAether is the card definition for Frozen Aether.
//
// Type: Enchantment
// Cost: {3}{U}
//
// Oracle text:
//
//	Artifacts, creatures, and lands your opponents control enter tapped.
var FrozenAether = newFrozenAether

func newFrozenAether() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Frozen Aether",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedGroupReplacement("Artifacts, creatures, and lands your opponents control enter tapped.", game.TriggerControllerOpponent, types.Artifact, types.Creature, types.Land),
			},
			OracleText: `
			Artifacts, creatures, and lands your opponents control enter tapped.
		`,
		},
	}
}
