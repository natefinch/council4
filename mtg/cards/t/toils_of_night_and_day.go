package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ToilsOfNightAndDay is the card definition for Toils of Night and Day.
//
// Type: Instant — Arcane
// Cost: {2}{U}
//
// Oracle text:
//
//	You may tap or untap target permanent, then you may tap or untap another target permanent.
var ToilsOfNightAndDay = newToilsOfNightAndDay()

func newToilsOfNightAndDay() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Toils of Night and Day",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Arcane},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target permanent",
						Allow:      game.TargetAllowPermanent,
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "another target permanent",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludeSource: true}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.TapOrUntap{
							Object: game.TargetPermanentReference(0),
						},
						Optional: true,
					},
					{
						Primitive: game.TapOrUntap{
							Object: game.TargetPermanentReference(1),
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			You may tap or untap target permanent, then you may tap or untap another target permanent.
		`,
		},
	}
}
