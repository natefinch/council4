package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GontiSMachinations is the card definition for Gonti's Machinations.
//
// Type: Enchantment
// Cost: {B}
//
// Oracle text:
//
//	Whenever you lose life for the first time each turn, you get {E}. (You get an energy counter. Damage causes loss of life.)
//	Pay {E}{E}, Sacrifice this enchantment: Each opponent loses 3 life. You gain life equal to the life lost this way.
var GontiSMachinations = newGontiSMachinations

func newGontiSMachinations() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gonti's Machinations",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Pay {E}{E}, Sacrifice this enchantment: Each opponent loses 3 life. You gain life equal to the life lost this way.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalEnergy,
							Text:        "Pay {E}{E}",
							Amount:      2,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this enchantment",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(3),
											PlayerGroup: game.OpponentsReference(),
										},
										PublishResult: game.ResultKey("life-change"),
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:      game.DynamicAmountPreviousEffectResult,
												ResultKey: game.ResultKey("life-change"),
											}),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you lose life for the first time each turn, you get {E}. (You get an energy counter. Damage causes loss of life.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventLifeLost,
							Player:                     game.TriggerPlayerYou,
							StackObjectKind:            game.StackSpell,
							PlayerEventOrdinalThisTurn: 1,
							CounterKind:                counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddPlayerCounter{
											Amount:      game.Fixed(1),
											Player:      game.ControllerReference(),
											CounterKind: counter.Energy,
										},
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
			Whenever you lose life for the first time each turn, you get {E}. (You get an energy counter. Damage causes loss of life.)
			Pay {E}{E}, Sacrifice this enchantment: Each opponent loses 3 life. You gain life equal to the life lost this way.
		`,
		},
	}
}
