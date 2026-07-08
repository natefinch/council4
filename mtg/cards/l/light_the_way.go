package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LightTheWay is the card definition for Light the Way.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Choose one —
//	• Put a +1/+1 counter on target creature or Vehicle. Untap it.
//	• Return target permanent you control to its owner's hand.
var LightTheWay = newLightTheWay()

func newLightTheWay() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Light the Way",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Put a +1/+1 counter on target creature or Vehicle. Untap it.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Creature}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}}}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Return target permanent you control to its owner's hand.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target permanent you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Put a +1/+1 counter on target creature or Vehicle. Untap it.
			• Return target permanent you control to its owner's hand.
		`,
		},
	}
}
