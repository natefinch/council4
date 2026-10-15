package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FrostBreath is the card definition for Frost Breath.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	Tap up to two target creatures. Those creatures don't untap during their controller's next untap step.
var FrostBreath = newFrostBreath

func newFrostBreath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Frost Breath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 2,
						Constraint: "up to two target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Tap{
							Object: game.TargetPermanentReference(1),
						},
					},
					{
						Primitive: game.SkipNextUntap{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.SkipNextUntap{
							Object: game.TargetPermanentReference(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap up to two target creatures. Those creatures don't untap during their controller's next untap step.
		`,
		},
	}
}
