package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HighspireInfusion is the card definition for Highspire Infusion.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Target creature gets +3/+3 until end of turn. You get {E}{E} (two energy counters).
var HighspireInfusion = newHighspireInfusion

func newHighspireInfusion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Highspire Infusion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
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
							ToughnessDelta: game.Fixed(3),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.AddPlayerCounter{
							Amount:      game.Fixed(2),
							Player:      game.ControllerReference(),
							CounterKind: counter.Energy,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets +3/+3 until end of turn. You get {E}{E} (two energy counters).
		`,
		},
	}
}
