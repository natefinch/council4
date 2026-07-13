package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AndúrilNarsilReforged is the card definition for Andúril, Narsil Reforged.
//
// Type: Legendary Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
//	Whenever equipped creature attacks, put a +1/+1 counter on each creature you control. If you have the city's blessing, put two +1/+1 counters on each creature you control instead.
//	Equip {3}
var AndúrilNarsilReforged = newAndúrilNarsilReforged

func newAndúrilNarsilReforged() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Andúril, Narsil Reforged",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.AscendStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Source:           game.TriggerSourceAttachedPermanent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									CounterKind: counter.PlusOnePlusOne,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										Negate:                    true,
										ControllerHasCityBlessing: true,
									}),
								}),
							},
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(2),
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									CounterKind: counter.PlusOnePlusOne,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										ControllerHasCityBlessing: true,
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
			Whenever equipped creature attacks, put a +1/+1 counter on each creature you control. If you have the city's blessing, put two +1/+1 counters on each creature you control instead.
			Equip {3}
		`,
		},
	}
}
