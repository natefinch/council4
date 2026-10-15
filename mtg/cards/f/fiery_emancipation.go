package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FieryEmancipation is the card definition for Fiery Emancipation.
//
// Type: Enchantment
// Cost: {3}{R}{R}{R}
//
// Oracle text:
//
//	If a source you control would deal damage to a permanent or player, it deals triple that damage to that permanent or player instead.
var FieryEmancipation = newFieryEmancipation

func newFieryEmancipation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fiery Emancipation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacement("If a source you control would deal damage to a permanent or player, it deals triple that damage to that permanent or player instead.", 3, 0, nil, game.TriggerControllerYou),
			},
			OracleText: `
			If a source you control would deal damage to a permanent or player, it deals triple that damage to that permanent or player instead.
		`,
		},
	}
}
