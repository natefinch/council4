package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GhastbarkTwins is the card definition for Ghastbark Twins.
//
// Type: Creature — Treefolk
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
//	This creature can block an additional creature each combat.
var GhastbarkTwins = newGhastbarkTwins()

func newGhastbarkTwins() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ghastbark Twins",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                 game.RuleEffectCanBlockAdditional,
							AffectedSource:       true,
							AdditionalBlockCount: 1,
						},
					},
				},
			},
			OracleText: `
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
			This creature can block an additional creature each combat.
		`,
		},
	}
}
