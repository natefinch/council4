package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MysteriousEgg is the card definition for Mysterious Egg.
//
// Type: Creature — Egg
// Cost: {1}
//
// Oracle text:
//
//	Whenever this creature mutates, put a +1/+1 counter on it.
var MysteriousEgg = newMysteriousEgg

func newMysteriousEgg() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mysterious Egg",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Egg},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature mutates, put a +1/+1 counter on it.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentMutated,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.EventPermanentReference(),
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
			Whenever this creature mutates, put a +1/+1 counter on it.
		`,
		},
	}
}
