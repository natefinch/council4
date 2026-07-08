package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AppliedGeometry is the card definition for Applied Geometry.
//
// Type: Sorcery
// Cost: {2}{G}{U}
//
// Oracle text:
//
//	Create a token that's a copy of target non-Aura permanent you control, except it's a 0/0 Fractal creature in addition to its other types. Put six +1/+1 counters on it.
var AppliedGeometry = newAppliedGeometry

func newAppliedGeometry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Applied Geometry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.U,
			}),
			Colors: []color.Color{color.Green, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target non-Aura permanent you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedSubtype: types.Sub("Aura"), Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source:       game.TokenCopySourceObject,
								Object:       game.TargetPermanentReference(0),
								SetPower:     opt.Val(game.PT{Value: 0}),
								SetToughness: opt.Val(game.PT{Value: 0}),
								AddTypes:     []types.Card{types.Creature},
								AddSubtypes:  []types.Sub{types.Fractal},
							}),
						},
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(6),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a token that's a copy of target non-Aura permanent you control, except it's a 0/0 Fractal creature in addition to its other types. Put six +1/+1 counters on it.
		`,
		},
	}
}
