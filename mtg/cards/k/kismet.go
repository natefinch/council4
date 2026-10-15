package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Kismet is the card definition for Kismet.
//
// Type: Enchantment
// Cost: {3}{W}
//
// Oracle text:
//
//	Artifacts, creatures, and lands your opponents control enter tapped.
var Kismet = newKismet

func newKismet() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Kismet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
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
