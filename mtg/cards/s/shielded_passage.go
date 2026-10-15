package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShieldedPassage is the card definition for Shielded Passage.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Prevent all damage that would be dealt to target creature this turn.
var ShieldedPassage = newShieldedPassage

func newShieldedPassage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Shielded Passage",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
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
						Primitive: game.PreventDamage{
							Object: game.TargetPermanentReference(0),
							All:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Prevent all damage that would be dealt to target creature this turn.
		`,
		},
	}
}
