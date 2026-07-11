package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Rouse is the card definition for Rouse.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	If you control a Swamp, you may pay 2 life rather than pay this spell's mana cost.
//	Target creature gets +2/+0 until end of turn.
var Rouse = newRouse

func newRouse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Rouse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Pay 2 life",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalPayLife,
							Text:   "pay 2 life",
							Amount: 2,
						},
					},
					Condition:        cost.AlternativeConditionControlsPermanentSubtype,
					ConditionSubtype: types.Swamp,
				},
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
							PowerDelta:     game.Fixed(2),
							ToughnessDelta: game.Fixed(0),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			If you control a Swamp, you may pay 2 life rather than pay this spell's mana cost.
			Target creature gets +2/+0 until end of turn.
		`,
		},
	}
}
