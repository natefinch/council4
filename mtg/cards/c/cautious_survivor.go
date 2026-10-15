package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CautiousSurvivor is the card definition for Cautious Survivor.
//
// Type: Creature — Elf Survivor
// Cost: {3}{G}
//
// Oracle text:
//
//	Survival — At the beginning of your second main phase, if this creature is tapped, you gain 2 life.
var CautiousSurvivor = newCautiousSurvivor

func newCautiousSurvivor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cautious Survivor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Survivor},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Survival — At the beginning of your second main phase, if this creature is tapped, you gain 2 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepPostcombatMain,
						},
						InterveningIf: "if this creature is tapped",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if this creature is tapped",
							Object:            opt.Val(game.SourcePermanentReference()),
							ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriTrue}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
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
			OracleText: `
			Survival — At the beginning of your second main phase, if this creature is tapped, you gain 2 life.
		`,
		},
	}
}
