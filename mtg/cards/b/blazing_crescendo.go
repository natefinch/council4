package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlazingCrescendo is the card definition for Blazing Crescendo.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Target creature gets +3/+1 until end of turn.
//	Exile the top card of your library. Until the end of your next turn, you may play that card.
var BlazingCrescendo = newBlazingCrescendo

func newBlazingCrescendo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Blazing Crescendo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
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
							PowerDelta:     game.Fixed(3),
							ToughnessDelta: game.Fixed(1),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ImpulseExile{
							Player:   game.ControllerReference(),
							Amount:   game.Fixed(1),
							Duration: game.DurationUntilEndOfYourNextTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets +3/+1 until end of turn.
			Exile the top card of your library. Until the end of your next turn, you may play that card.
		`,
		},
	}
}
