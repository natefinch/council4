package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CloakedCadet is the card definition for Cloaked Cadet.
//
// Type: Creature — Human Ranger
// Cost: {4}{G}
//
// Oracle text:
//
//	Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
//	Whenever one or more +1/+1 counters are put on one or more Humans you control, draw a card. This ability triggers only once each turn.
var CloakedCadet = newCloakedCadet

func newCloakedCadet() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cloaked Cadet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Ranger},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TrainingTriggeredBody,
				game.TriggeredAbility{
					Text: "Whenever one or more +1/+1 counters are put on one or more Humans you control, draw a card. This ability triggers only once each turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventCountersAdded,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Human}},
							StackObjectKind:  game.StackSpell,
							OneOrMore:        true,
							MatchCounterKind: true,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
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
			Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
			Whenever one or more +1/+1 counters are put on one or more Humans you control, draw a card. This ability triggers only once each turn.
		`,
		},
	}
}
