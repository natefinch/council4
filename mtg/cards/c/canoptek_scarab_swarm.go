package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CanoptekScarabSwarm is the card definition for Canoptek Scarab Swarm.
//
// Type: Artifact Creature — Insect
// Cost: {4}
//
// Oracle text:
//
//	Flying
//	Feeder Mandibles — When this creature enters, exile target player's graveyard. For each artifact or land card exiled this way, create a 1/1 colorless Insect artifact creature token with flying.
var CanoptekScarabSwarm = newCanoptekScarabSwarm

func newCanoptekScarabSwarm() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Canoptek Scarab Swarm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Feeder Mandibles — When this creature enters, exile target player's graveyard. For each artifact or land card exiled this way, create a 1/1 colorless Insect artifact creature token with flying.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target player's graveyard",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Player:      game.TargetPlayerReference(0),
											FromZone:    zone.Graveyard,
											Destination: zone.Exile,
										},
										PublishResult: game.ResultKey("actual-card-quantity-1"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("actual-card-quantity-1")},
										},
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountPreviousEffectResult,
												Multiplier: 1,
												Selection:  &game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Land}},
												ResultKey:  game.ResultKey("actual-card-quantity-1"),
											}),
											Source: game.TokenDef(canoptekScarabSwarmToken),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:      game.ResultKey("actual-card-quantity-1"),
											Accepted: game.TriTrue,
										}),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Flying
			Feeder Mandibles — When this creature enters, exile target player's graveyard. For each artifact or land card exiled this way, create a 1/1 colorless Insect artifact creature token with flying.
		`,
		},
	}
}

var canoptekScarabSwarmToken = newCanoptekScarabSwarmToken()

func newCanoptekScarabSwarmToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Insect",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
