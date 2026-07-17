package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EntrancingMelody is the card definition for Entrancing Melody.
//
// Type: Sorcery
// Cost: {X}{U}{U}
//
// Oracle text:
//
//	Gain control of target creature with mana value X.
var EntrancingMelody = newEntrancingMelody

func newEntrancingMelody() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Entrancing Melody",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:       1,
						MaxTargets:       1,
						Constraint:       "target creature with mana value X",
						Allow:            game.TargetAllowPermanent,
						Selection:        opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
						ManaValueEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:         game.LayerControl,
									NewController: opt.Val(game.Player1),
								},
							},
							Duration: game.DurationPermanent,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Gain control of target creature with mana value X.
		`,
		},
	}
}
