// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTBlinkedSubject is the card definition for RT Blinked Subject.
//
// Type: Instant
//
// Oracle text:
//
//	Exile target creature you control, then return it to the battlefield under its owner's control. If that creature is a Bird, Frog, Otter, or Rat, draw a card.
var RTBlinkedSubject = newRTBlinkedSubject

func newRTBlinkedSubject() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Blinked Subject",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:        game.TargetPermanentReference(0),
							Destination:   zone.Exile,
							PublishLinked: game.LinkedKey("blink-1"),
						},
					},
					{
						Primitive: game.PutOnBattlefield{
							Source:        game.LinkedBattlefieldSource(game.LinkedKey("blink-1")),
							PublishLinked: game.LinkedKey("sequence-effect-1-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-1-product")},
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.LinkedObjectReference("sequence-effect-1-product")),
								ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Bird"), types.Sub("Frog"), types.Sub("Otter"), types.Sub("Rat")}}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Exile target creature you control, then return it to the battlefield under its owner's control. If that creature is a Bird, Frog, Otter, or Rat, draw a card.
		`,
		},
	}
}
