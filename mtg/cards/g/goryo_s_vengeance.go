package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GoryoSVengeance is the card definition for Goryo's Vengeance.
//
// Type: Instant — Arcane
// Cost: {1}{B}
//
// Oracle text:
//
//	Return target legendary creature card from your graveyard to the battlefield. That creature gains haste. Exile it at the beginning of the next end step.
//	Splice onto Arcane {2}{B} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
var GoryoSVengeance = newGoryoSVengeance

func newGoryoSVengeance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Goryo's Vengeance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Arcane},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SpliceKeyword{Cost: cost.Mana{cost.O(2), cost.B}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target legendary creature card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							PublishLinked: game.LinkedKey("gain-keyword-1"),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.LinkedObjectReference("gain-keyword-1")),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									AddKeywords: []game.Keyword{
										game.Haste,
									},
								},
							},
							Duration:      game.DurationPermanent,
							PublishLinked: game.LinkedKey("delayed-exile-2"),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing:         game.DelayedAtBeginningOfNextEndStep,
								CapturedObject: opt.Val(game.LinkedObjectReference("delayed-exile-2")),
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.Exile{
												Object: game.CapturedObjectReference(),
											},
										},
									},
								}.Ability(),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target legendary creature card from your graveyard to the battlefield. That creature gains haste. Exile it at the beginning of the next end step.
			Splice onto Arcane {2}{B} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
		`,
		},
	}
}
