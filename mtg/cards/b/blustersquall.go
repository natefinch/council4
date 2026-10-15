package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Blustersquall is the card definition for Blustersquall.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Tap target creature you don't control.
//	Overload {3}{U} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")
var Blustersquall = newBlustersquall

func newBlustersquall() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Blustersquall",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			Overload: opt.Val(game.OverloadAbility{Cost: cost.Mana{cost.O(3), cost.U}, SpellAbility: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerNotYou}),
						},
					},
				},
			}.Ability()}),
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you don't control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerNotYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap target creature you don't control.
			Overload {3}{U} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")
		`,
		},
	}
}
