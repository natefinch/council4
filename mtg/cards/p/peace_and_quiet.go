package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PeaceAndQuiet is the card definition for Peace and Quiet.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Destroy two target enchantments.
var PeaceAndQuiet = newPeaceAndQuiet

func newPeaceAndQuiet() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Peace and Quiet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 2,
						MaxTargets: 2,
						Constraint: "two target enchantments",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy two target enchantments.
		`,
		},
	}
}
