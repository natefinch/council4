package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AngrathSMarauders is the card definition for Angrath's Marauders.
//
// Type: Creature — Human Pirate
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	If a source you control would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.
var AngrathSMarauders = newAngrathSMarauders

func newAngrathSMarauders() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Angrath's Marauders",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Pirate},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamageReplacement("If a source you control would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.", 2, 0, nil, game.TriggerControllerYou),
			},
			OracleText: `
			If a source you control would deal damage to a permanent or player, it deals double that damage to that permanent or player instead.
		`,
		},
	}
}
