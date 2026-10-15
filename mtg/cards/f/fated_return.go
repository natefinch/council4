package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FatedReturn is the card definition for Fated Return.
//
// Type: Instant
// Cost: {4}{B}{B}{B}
//
// Oracle text:
//
//	Put target creature card from a graveyard onto the battlefield under your control. It gains indestructible. If it's your turn, scry 2.
var FatedReturn = newFatedReturn

func newFatedReturn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Fated Return",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card from a graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							Recipient:     opt.Val(game.ControllerReference()),
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.Indestructible},
								},
							},
							Duration: game.DurationPermanent,
						},
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SourceControllerTurn: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Put target creature card from a graveyard onto the battlefield under your control. It gains indestructible. If it's your turn, scry 2.
		`,
		},
	}
}
