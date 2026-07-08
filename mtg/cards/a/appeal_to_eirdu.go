package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AppealToEirdu is the card definition for Appeal to Eirdu.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	One or two target creatures each get +2/+1 until end of turn.
var AppealToEirdu = newAppealToEirdu

func newAppealToEirdu() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Appeal to Eirdu",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 2,
						Constraint: "one or two target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(1),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(1),
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(1),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			One or two target creatures each get +2/+1 until end of turn.
		`,
		},
	}
}
