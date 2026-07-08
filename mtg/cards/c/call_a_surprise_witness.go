package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CallASurpriseWitness is the card definition for Call a Surprise Witness.
//
// Type: Sorcery
// Cost: {1}{W}
//
// Oracle text:
//
//	Return target creature card with mana value 3 or less from your graveyard to the battlefield. Put a flying counter on it. It's a Spirit in addition to its other types.
var CallASurpriseWitness = newCallASurpriseWitness

func newCallASurpriseWitness() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Call a Surprise Witness",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card with mana value 3 or less from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							PublishLinked: game.LinkedKey("leave-bf-exile-1"),
						},
					},
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(1),
							Object:      game.LinkedObjectReference("leave-bf-exile-1"),
							CounterKind: counter.Flying,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.LinkedObjectReference("leave-bf-exile-1")),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:       game.LayerType,
									AddSubtypes: []types.Sub{types.Sub("Spirit")},
								},
							},
							Duration: game.DurationPermanent,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target creature card with mana value 3 or less from your graveyard to the battlefield. Put a flying counter on it. It's a Spirit in addition to its other types.
		`,
		},
	}
}
