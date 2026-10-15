package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ScourFromExistence is the card definition for Scour from Existence.
//
// Type: Instant
// Cost: {7}
//
// Oracle text:
//
//	Exile target permanent.
var ScourFromExistence = newScourFromExistence

func newScourFromExistence() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Scour from Existence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target permanent",
						Allow:      game.TargetAllowPermanent,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Exile target permanent.
		`,
		},
	}
}
