package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wolf is the card definition for Wolf.
//
// Type: Token Creature — Wolf
//
// Oracle text:
//   When this creature dies, put a loyalty counter on each Garruk you control.

// WolfToken21c3b763b6a24e9e8621d2ab62e448f2 is the card definition for Wolf.
var WolfToken21c3b763b6a24e9e8621d2ab62e448f2 = newWolfToken21c3b763b6a24e9e8621d2ab62e448f2()

func newWolfToken21c3b763b6a24e9e8621d2ab62e448f2() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:      "Wolf",
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature dies, put a loyalty counter on each Garruk you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Group:       game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Garruk")}, Controller: game.ControllerYou}),
											CounterKind: counter.Loyalty,
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
			When this creature dies, put a loyalty counter on each Garruk you control.
		`,
		},
	}
}
