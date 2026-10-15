package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FurnaceOfRath is the card definition for Furnace of Rath.
//
// Type: Enchantment
// Cost: {1}{R}{R}{R}
//
// Oracle text:
//
//	If a source would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.
var FurnaceOfRath = newFurnaceOfRath

func newFurnaceOfRath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Furnace of Rath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacementFiltered("If a source would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.", &game.DamageReplacementSpec{Multiplier: 2, Addend: 0, SourceColors: nil, SourceTypes: nil, ExcludeSource: false, RecipientOpponent: false, RecipientOpponentPlayerOnly: false, NoncombatOnly: false, Controller: game.TriggerControllerAny}),
			},
			OracleText: `
			If a source would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.
		`,
		},
	}
}
