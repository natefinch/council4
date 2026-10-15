package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GratuitousViolence is the card definition for Gratuitous Violence.
//
// Type: Enchantment
// Cost: {2}{R}{R}{R}
//
// Oracle text:
//
//	If a creature you control would deal damage to a permanent or player, it deals double that damage instead.
var GratuitousViolence = newGratuitousViolence

func newGratuitousViolence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Gratuitous Violence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacementFiltered("If a creature you control would deal damage to a permanent or player, it deals double that damage instead.", &game.DamageReplacementSpec{Multiplier: 2, Addend: 0, SourceColors: nil, SourceTypes: []types.Card{types.Creature}, ExcludeSource: false, RecipientOpponent: false, RecipientOpponentPlayerOnly: false, NoncombatOnly: false, Controller: game.TriggerControllerYou}),
			},
			OracleText: `
			If a creature you control would deal damage to a permanent or player, it deals double that damage instead.
		`,
		},
	}
}
