package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmbermawHellion is the card definition for Embermaw Hellion.
//
// Type: Creature — Hellion
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.
var EmbermawHellion = newEmbermawHellion

func newEmbermawHellion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Embermaw Hellion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hellion},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacementExcludingSource("If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.", 0, 1, []color.Color{color.Red}, game.TriggerControllerYou),
			},
			OracleText: `
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
			If another red source you control would deal damage to a permanent or player, it deals that much damage plus 1 to that permanent or player instead.
		`,
		},
	}
}
