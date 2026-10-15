package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SheoldredSRestoration is the card definition for Sheoldred's Restoration.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	Kicker {2}{W} (You may pay an additional {2}{W} as you cast this spell.)
//	Return target creature card from your graveyard to the battlefield. If this spell was kicked, you gain life equal to that card's mana value. Otherwise, you lose that much life.
//	Exile Sheoldred's Restoration.
var SheoldredSRestoration = newSheoldredSRestoration

func newSheoldredSRestoration() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Sheoldred's Restoration",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Kicker {2}{W}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(2),
								cost.W,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.PutOnBattlefield{
							Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
						},
						PublishResult: game.ResultKey("life-rider-move-1"),
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountObjectManaValue,
								Multiplier: 1,
								Object:     game.LinkedObjectReference("sequence-effect-0-product"),
							}),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellWasKicked: true,
							}),
						}),
						PublishCondition: "condition-0",
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "life-rider-move-1",
							Succeeded: game.TriTrue,
						}),
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountObjectManaValue,
								Multiplier: 1,
								Object:     game.LinkedObjectReference("sequence-effect-0-product"),
							}),
							Player: game.ControllerReference(),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "life-rider-move-1",
							Succeeded: game.TriTrue,
						}),
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Exile,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Kicker {2}{W} (You may pay an additional {2}{W} as you cast this spell.)
			Return target creature card from your graveyard to the battlefield. If this spell was kicked, you gain life equal to that card's mana value. Otherwise, you lose that much life.
			Exile Sheoldred's Restoration.
		`,
		},
	}
}
