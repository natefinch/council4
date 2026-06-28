package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EchoingCalm is the card definition for Echoing Calm.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Destroy target enchantment and all other enchantments with the same name as that enchantment.
var EchoingCalm = newEchoingCalm()

func newEchoingCalm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Echoing Calm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target enchantment and all other enchantments with the same name as that enchantment",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target enchantment and all other enchantments with the same name as that enchantment.
		`,
		},
	}
}
