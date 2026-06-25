package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Decommission is the card definition for Decommission.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Destroy target artifact or enchantment.
//	Revolt — If a permanent left the battlefield under your control this turn, you gain 3 life.
var Decommission = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name: "Decommission",
		ManaCost: opt.Val(cost.Mana{
			cost.O(2),
			cost.W,
		}),
		Colors: []color.Color{color.White},
		Types:  []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target artifact or enchantment",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
				},
			},
			Sequence: []game.Instruction{
				{
					Primitive: game.Destroy{
						Object: game.TargetPermanentReference(0),
					},
				},
				{
					Primitive: game.GainLife{
						Amount: game.Fixed(3),
						Player: game.ControllerReference(),
					},
					Condition: opt.Val(game.EffectCondition{
						Condition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:         game.EventZoneChanged,
								Controller:    game.TriggerControllerYou,
								MatchFromZone: true,
								FromZone:      zone.Battlefield,
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					}),
				},
			},
		}.Ability()),
		OracleText: `
			Destroy target artifact or enchantment.
			Revolt — If a permanent left the battlefield under your control this turn, you gain 3 life.
		`,
	},
}
