package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MortalCombat is the card definition for Mortal Combat.
//
// Type: Enchantment
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	At the beginning of your upkeep, if twenty or more creature cards are in your graveyard, you win the game.
var MortalCombat = newMortalCombat

func newMortalCombat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Mortal Combat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, if twenty or more creature cards are in your graveyard, you win the game.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
						InterveningIf: "if twenty or more creature cards are in your graveyard",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if twenty or more creature cards are in your graveyard",
							SourceCounterKind: counter.PlusOnePlusOne,
							ControllerGraveyardCardOfTypeCountAtLeast: 20,
							ControllerGraveyardCountCardType:          types.Creature,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PlayerWinsGame{
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
			At the beginning of your upkeep, if twenty or more creature cards are in your graveyard, you win the game.
		`,
		},
	}
}
