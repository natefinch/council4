// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTModalSubject is the card definition for RT Modal Subject.
//
// Type: Instant
//
// Oracle text:
//
//	Choose one —
//	• Exile target creature you control, then return it to the battlefield tapped under its owner's control. If it's an Elf, untap it.
//	• Draw a card.
var RTModalSubject = newRTModalSubject

func newRTModalSubject() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Modal Subject",
			Types: []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Exile target creature you control, then return it to the battlefield tapped under its owner's control. If it's an Elf, untap it.",
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
									EntryTapped:   true,
									PublishLinked: game.LinkedKey("sequence-effect-1-product"),
								},
								LocalProducts: game.LocalProducts{
									Links: []game.LinkedKey{game.LinkedKey("sequence-effect-1-product")},
								},
							},
							{
								Primitive: game.Untap{
									Object: game.LinkedObjectReference("sequence-effect-1-product"),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Object:        opt.Val(game.LinkedObjectReference("sequence-effect-1-product")),
										ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}}),
									}),
								}),
							},
						},
					},
					game.Mode{
						Text: "Draw a card.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
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
			• Exile target creature you control, then return it to the battlefield tapped under its owner's control. If it's an Elf, untap it.
			• Draw a card.
		`,
		},
	}
}
