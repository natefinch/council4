package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HeartShapedHerb is the card definition for Heart-Shaped Herb.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	If a source an opponent controls would deal damage to you, prevent 1 of that damage.
//	{2}, {T}, Sacrifice this artifact: You may sacrifice a creature. If you do, return that card to the battlefield under its owner's control with three +1/+1 counters on it and you become the monarch.
var HeartShapedHerb = newHeartShapedHerb()

func newHeartShapedHerb() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Heart-Shaped Herb",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, {T}, Sacrifice this artifact: You may sacrifice a creature. If you do, return that card to the battlefield under its owner's control with three +1/+1 counters on it and you become the monarch.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this artifact",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:        game.Fixed(1),
									Player:        game.ControllerReference(),
									Selection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
									PublishLinked: game.LinkedKey("sacrificed-creature"),
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.PutOnBattlefield{
									Source:            game.LinkedBattlefieldSource(game.LinkedKey("sacrificed-creature")),
									EntryCounters:     []game.CounterPlacement{game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 3}},
									LinkedReturnZones: []zone.Type{zone.Graveyard},
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionReplacement("If a source an opponent controls would deal damage to you, prevent 1 of that damage.", &game.DamagePreventionSpec{Amount: 1, SourceColors: nil, SourceTypes: nil, SourceControllerOpponent: true}),
			},
			OracleText: `
			If a source an opponent controls would deal damage to you, prevent 1 of that damage.
			{2}, {T}, Sacrifice this artifact: You may sacrifice a creature. If you do, return that card to the battlefield under its owner's control with three +1/+1 counters on it and you become the monarch.
		`,
		},
	}
}
