// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTReachedTargetCard is the card definition for RT Reached Target Card.
//
// Type: Creature
//
// Oracle text:
//
//	{1}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this creature.
var RTReachedTargetCard = newRTReachedTargetCard

func newRTReachedTargetCard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "RT Reached Target Card",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target card from a graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Exile,
										},
										PublishResult: game.ResultKey("condition-target-card-1"),
									},
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                "If it was a permanent card",
												Object:              opt.Val(game.TargetCardReference(0)),
												ObjectMatches:       opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}}),
												TargetCardResultKey: game.ResultKey("condition-target-card-1"),
												SourceCounterKind:   counter.PlusOnePlusOne,
											}),
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
			{1}: Exile target card from a graveyard. If it was a permanent card, put a +1/+1 counter on this creature.
		`,
		},
	}
}
