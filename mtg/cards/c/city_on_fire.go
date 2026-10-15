package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CityOnFire is the card definition for City on Fire.
//
// Type: Enchantment
// Cost: {5}{R}{R}{R}
//
// Oracle text:
//
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	If a source you control would deal damage to a permanent or player, it deals triple that damage instead.
var CityOnFire = newCityOnFire

func newCityOnFire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "City on Fire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacement("If a source you control would deal damage to a permanent or player, it deals triple that damage instead.", 3, 0, nil, game.TriggerControllerYou),
			},
			OracleText: `
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			If a source you control would deal damage to a permanent or player, it deals triple that damage instead.
		`,
		},
	}
}
