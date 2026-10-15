package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SuddenDeath is the card definition for Sudden Death.
//
// Type: Instant
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Split second (As long as this spell is on the stack, players can't cast spells or activate abilities that aren't mana abilities.)
//	Target creature gets -4/-4 until end of turn.
var SuddenDeath = newSuddenDeath

func newSuddenDeath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sudden Death",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.SplitSecondStaticBody,
			},
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
							PowerDelta:     game.Fixed(-4),
							ToughnessDelta: game.Fixed(-4),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Split second (As long as this spell is on the stack, players can't cast spells or activate abilities that aren't mana abilities.)
			Target creature gets -4/-4 until end of turn.
		`,
		},
	}
}
