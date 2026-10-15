package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WillOfTheSultai is the card definition for Will of the Sultai.
//
// Type: Sorcery
// Cost: {4}{G}
//
// Oracle text:
//
//	Choose one. If you control a commander as you cast this spell, you may choose both instead.
//	• Target player mills three cards. Return all land cards from your graveyard to the battlefield tapped.
//	• Put X +1/+1 counters on target creature, where X is the number of lands you control. It gains trample until end of turn.
var WillOfTheSultai = newWillOfTheSultai

func newWillOfTheSultai() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Will of the Sultai",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target player mills three cards. Return all land cards from your graveyard to the battlefield tapped.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveTopOfLibrary{
									Amount:      game.Fixed(3),
									Player:      game.TargetPlayerReference(0),
									Destination: zone.Graveyard,
								},
							},
							{
								Primitive: game.MassReturnFromGraveyard{
									Player:      game.ControllerReference(),
									Selection:   game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou},
									Destination: zone.Battlefield,
									EntryTapped: true,
								},
							},
						},
					},
					game.Mode{
						Text: "Put X +1/+1 counters on target creature, where X is the number of lands you control. It gains trample until end of turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
									}),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:    game.DurationPermanent,
											Layer:       game.LayerAbility,
											AddKeywords: []game.Keyword{game.Trample},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
				},
				MinModes:        1,
				MaxModes:        1,
				ModeChoiceBonus: game.ModeChoiceBonus{Condition: game.ModeChoiceConditionControlsCommander, AdditionalMaxModes: 1},
			}),
			OracleText: `
			Choose one. If you control a commander as you cast this spell, you may choose both instead.
			• Target player mills three cards. Return all land cards from your graveyard to the battlefield tapped.
			• Put X +1/+1 counters on target creature, where X is the number of lands you control. It gains trample until end of turn.
		`,
		},
	}
}
