package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Trueheart Duelist
//
// Type: Token Creature — Zombie Human Warrior
//
// Oracle text:
//   Trueheart Duelist can block an additional creature each combat.

// TrueheartDuelistToken1cb54b3e92cb4b56a726468c89eb29a0 is the card definition for Trueheart Duelist.
var TrueheartDuelistToken1cb54b3e92cb4b56a726468c89eb29a0 = newTrueheartDuelistToken1cb54b3e92cb4b56a726468c89eb29a0()

func newTrueheartDuelistToken1cb54b3e92cb4b56a726468c89eb29a0() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Trueheart Duelist",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
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
			Trueheart Duelist can block an additional creature each combat.
		`,
		},
	}
}
