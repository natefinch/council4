package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CyclonicRift is the card definition for Cyclonic Rift.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Return target nonland permanent you don't control to its owner's hand.
//	Overload {6}{U} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")
var CyclonicRift = newCyclonicRift

func newCyclonicRift() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Cyclonic Rift",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			Overload: opt.Val(game.OverloadAbility{Cost: cost.Mana{cost.O(6), cost.U}, SpellAbility: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Group:       game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerNotYou}),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()}),
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonland permanent you don't control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerNotYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target nonland permanent you don't control to its owner's hand.
			Overload {6}{U} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")
		`,
		},
	}
}
