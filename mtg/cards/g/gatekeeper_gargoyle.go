package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GatekeeperGargoyle is the card definition for Gatekeeper Gargoyle.
//
// Type: Artifact Creature — Gargoyle
// Cost: {6}
//
// Oracle text:
//
//	Flying
//	This creature enters with a +1/+1 counter on it for each Gate you control.
var GatekeeperGargoyle = newGatekeeperGargoyle

func newGatekeeperGargoyle() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Gatekeeper Gargoyle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a +1/+1 counter on it for each Gate you control.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountCountSelector,
					Multiplier: 1,
					Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Gate")}, Controller: game.ControllerYou}),
				})}),
			},
			OracleText: `
			Flying
			This creature enters with a +1/+1 counter on it for each Gate you control.
		`,
		},
	}
}
