package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DictateOfTheTwinGods is the card definition for Dictate of the Twin Gods.
//
// Type: Enchantment
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Flash
//	If a source would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.
var DictateOfTheTwinGods = newDictateOfTheTwinGods

func newDictateOfTheTwinGods() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Dictate of the Twin Gods",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacementFiltered("If a source would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.", &game.DamageReplacementSpec{Multiplier: 2, Addend: 0, SourceColors: nil, SourceTypes: nil, ExcludeSource: false, RecipientOpponent: false, RecipientOpponentPlayerOnly: false, NoncombatOnly: false, Controller: game.TriggerControllerAny}),
			},
			OracleText: `
			Flash
			If a source would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.
		`,
		},
	}
}
