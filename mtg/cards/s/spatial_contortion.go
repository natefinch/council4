package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpatialContortion is the card definition for Spatial Contortion.
//
// Type: Instant
// Cost: {1}{C}
//
// Oracle text:
//
//	({C} represents colorless mana.)
//	Target creature gets +3/-3 until end of turn.
var SpatialContortion = newSpatialContortion

func newSpatialContortion() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Spatial Contortion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.C,
			}),
			Types: []types.Card{types.Instant},
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
							ToughnessDelta: game.Fixed(-3),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			({C} represents colorless mana.)
			Target creature gets +3/-3 until end of turn.
		`,
		},
	}
}
