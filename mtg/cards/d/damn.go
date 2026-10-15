package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Damn is the card definition for Damn.
//
// Type: Sorcery
// Cost: {B}{B}
//
// Oracle text:
//
//	Destroy target creature. A creature destroyed this way can't be regenerated.
//	Overload {2}{W}{W} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")
var Damn = newDamn

func newDamn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Damn",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			Overload: opt.Val(game.OverloadAbility{Cost: cost.Mana{cost.O(2), cost.W, cost.W}, SpellAbility: game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group:               game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							PreventRegeneration: true,
						},
					},
				},
			}.Ability()}),
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object:              game.TargetPermanentReference(0),
							PreventRegeneration: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature. A creature destroyed this way can't be regenerated.
			Overload {2}{W}{W} (You may cast this spell for its overload cost. If you do, change "target" in its text to "each.")
		`,
		},
	}
}
