package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RecklessImp is the card definition for Reckless Imp.
//
// Type: Creature — Imp
// Cost: {2}{B}
//
// Oracle text:
//
//	Flying
//	This creature can't block.
//	Dash {1}{B} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
var RecklessImp = newRecklessImp

func newRecklessImp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Reckless Imp",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Imp},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.CantBlockStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.DashTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Dash",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
					Mechanic: cost.AlternativeMechanicDash,
				},
			},
			OracleText: `
			Flying
			This creature can't block.
			Dash {1}{B} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
		`,
		},
	}
}
