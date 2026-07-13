package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VexingBauble is the card definition for Vexing Bauble.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	Whenever a player casts a spell, if no mana was spent to cast it, counter that spell.
//	{1}, {T}, Sacrifice this artifact: Draw a card.
var VexingBauble = newVexingBauble

func newVexingBauble() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Vexing Bauble",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, {T}, Sacrifice this artifact: Draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
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
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event: game.EventSpellCast,
						},
						InterveningIf: "if no mana was spent to cast it",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateEventSpellManaSpentToCast, Op: compare.LessOrEqual, Value: 0}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CounterObject{
									Object: game.EventStackObjectReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a player casts a spell, if no mana was spent to cast it, counter that spell.
			{1}, {T}, Sacrifice this artifact: Draw a card.
		`,
		},
	}
}
