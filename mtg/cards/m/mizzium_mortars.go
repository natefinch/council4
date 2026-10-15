package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MizziumMortars is the card definition for Mizzium Mortars.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Mizzium Mortars deals 4 damage to target creature you don't control.
//	Overload {3}{R}{R}{R} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")
var MizziumMortars = newMizziumMortars

func newMizziumMortars() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mizzium Mortars",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			Overload: opt.Val(game.OverloadAbility{Cost: cost.Mana{cost.O(3), cost.R, cost.R, cost.R}, SpellAbility: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerNotYou})),
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
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Mizzium Mortars deals 4 damage to target creature you don't control.
			Overload {3}{R}{R}{R} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")
		`,
		},
	}
}
