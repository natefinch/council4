package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GixianSkullflayer is the card definition for Gixian Skullflayer.
//
// Type: Creature — Phyrexian Human Assassin
// Cost: {2}{B}
//
// Oracle text:
//
//	At the beginning of your upkeep, if there are three or more creature cards in your graveyard, put a +1/+1 counter on this creature.
var GixianSkullflayer = newGixianSkullflayer

func newGixianSkullflayer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gixian Skullflayer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Human, types.Assassin},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, if there are three or more creature cards in your graveyard, put a +1/+1 counter on this creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
						InterveningIf: "if there are three or more creature cards in your graveyard",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if there are three or more creature cards in your graveyard",
							SourceCounterKind: counter.PlusOnePlusOne,
							ControllerGraveyardCardOfTypeCountAtLeast: 3,
							ControllerGraveyardCountCardType:          types.Creature,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
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
			At the beginning of your upkeep, if there are three or more creature cards in your graveyard, put a +1/+1 counter on this creature.
		`,
		},
	}
}
