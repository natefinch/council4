package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ExcaliburII is the card definition for Excalibur II.
//
// Type: Legendary Artifact — Equipment
// Cost: {1}
//
// Oracle text:
//
//	Whenever you gain life, put a charge counter on Excalibur II.
//	Equipped creature gets +1/+1 for each charge counter on Excalibur II.
//	Equip {3}
var ExcaliburII = newExcaliburII()

func newExcaliburII() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Excalibur II",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerPowerToughnessModify,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								Multiplier:  1,
								CounterKind: counter.Charge,
								Object:      game.SourcePermanentReference(),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountObjectCounters,
								Multiplier:  1,
								CounterKind: counter.Charge,
								Object:      game.SourcePermanentReference(),
							}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventLifeGained,
							Player: game.TriggerPlayerYou,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Charge,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you gain life, put a charge counter on Excalibur II.
			Equipped creature gets +1/+1 for each charge counter on Excalibur II.
			Equip {3}
		`,
		},
	}
}
