package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dismember is the card definition for Dismember.
//
// Type: Instant
// Cost: {1}{B/P}{B/P}
//
// Oracle text:
//
//	({B/P} can be paid with either {B} or 2 life.)
//	Target creature gets -5/-5 until end of turn.
var Dismember = newDismember

func newDismember() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dismember",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.PhyrexianMana(mana.B),
				cost.PhyrexianMana(mana.B),
			}),
			Colors: []color.Color{color.Black},
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
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(-5),
							ToughnessDelta: game.Fixed(-5),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			({B/P} can be paid with either {B} or 2 life.)
			Target creature gets -5/-5 until end of turn.
		`,
		},
	}
}
