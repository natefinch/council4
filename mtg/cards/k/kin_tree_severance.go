package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KinTreeSeverance is the card definition for Kin-Tree Severance.
//
// Type: Instant
// Cost: {2/W}{2/B}{2/G}
//
// Oracle text:
//
//	Exile target permanent with mana value 3 or greater.
var KinTreeSeverance = newKinTreeSeverance

func newKinTreeSeverance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Kin-Tree Severance",
			ManaCost: opt.Val(cost.Mana{
				cost.Twobrid(mana.W),
				cost.Twobrid(mana.B),
				cost.Twobrid(mana.G),
			}),
			Colors: []color.Color{color.Black, color.Green, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target permanent with mana value 3 or greater",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 3})}),
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
			Exile target permanent with mana value 3 or greater.
		`,
		},
	}
}
