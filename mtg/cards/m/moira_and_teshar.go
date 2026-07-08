package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MoiraAndTeshar is the card definition for Moira and Teshar.
//
// Type: Legendary Creature — Phyrexian Spirit Bird
// Cost: {3}{W}{B}
//
// Oracle text:
//
//	Flying
//	Whenever you cast a historic spell, return target nonland permanent card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step. If it would leave the battlefield, exile it instead of putting it anywhere else. (Artifacts, legendaries, and Sagas are historic.)
var MoiraAndTeshar = newMoiraAndTeshar

func newMoiraAndTeshar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Moira and Teshar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Spirit, types.Bird},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							RequireHistoric: true,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target nonland permanent card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Enchantment, types.Land, types.Planeswalker, types.Battle}, ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
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
										Timing: game.DelayedAtBeginningOfNextEndStep,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Exile{
														Object: game.LinkedObjectReference("delayed-exile-2"),
													},
												},
											},
										}.Ability(),
									},
								},
							},
							{
								Primitive: game.CreateReplacement{
									Replacement: &game.ReplacementEffect{
										MatchEvent:    game.EventZoneChanged,
										MatchFromZone: true,
										FromZone:      zone.Battlefield,
										ReplaceToZone: zone.Exile,
									},
									Object: game.LinkedObjectReference("delayed-exile-2"),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever you cast a historic spell, return target nonland permanent card from your graveyard to the battlefield. It gains haste. Exile it at the beginning of the next end step. If it would leave the battlefield, exile it instead of putting it anywhere else. (Artifacts, legendaries, and Sagas are historic.)
		`,
		},
	}
}
