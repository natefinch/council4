package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BattleSquadron is the card definition for Battle Squadron.
//
// Type: Creature — Goblin
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Flying
//	Battle Squadron's power and toughness are each equal to the number of creatures you control.
var BattleSquadron = newBattleSquadron

func newBattleSquadron() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Battle Squadron",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:           []color.Color{color.Red},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Goblin},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
			Battle Squadron's power and toughness are each equal to the number of creatures you control.
		`,
		},
	}
}
