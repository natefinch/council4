package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ScrollOfTheMasters is the card definition for Scroll of the Masters.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Whenever you cast a noncreature spell, put a lore counter on this artifact.
//	{3}, {T}: Target creature you control gets +1/+1 until end of turn for each lore counter on this artifact.
var ScrollOfTheMasters = newScrollOfTheMasters

func newScrollOfTheMasters() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Scroll of the Masters",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{3}, {T}: Target creature you control gets +1/+1 until end of turn for each lore counter on this artifact.",
					ManaCost:        opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object: game.TargetPermanentReference(0),
									PowerDelta: game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										Multiplier:  1,
										CounterKind: counter.Lore,
										Object:      game.SourcePermanentReference(),
									}),
									ToughnessDelta: game.Dynamic(game.DynamicAmount{
										Kind:        game.DynamicAmountObjectCounters,
										Multiplier:  1,
										CounterKind: counter.Lore,
										Object:      game.SourcePermanentReference(),
									}),
									Duration: game.DurationUntilEndOfTurn,
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
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Lore,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you cast a noncreature spell, put a lore counter on this artifact.
			{3}, {T}: Target creature you control gets +1/+1 until end of turn for each lore counter on this artifact.
		`,
		},
	}
}
