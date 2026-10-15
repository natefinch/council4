package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BerserkMurlodont is the card definition for Berserk Murlodont.
//
// Type: Creature — Beast
// Cost: {4}{G}
//
// Oracle text:
//
//	Whenever a Beast becomes blocked, it gets +1/+1 until end of turn for each creature blocking it.
var BerserkMurlodont = newBerserkMurlodont

func newBerserkMurlodont() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Berserk Murlodont",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a Beast becomes blocked, it gets +1/+1 until end of turn for each creature blocking it.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerBecameBlocked,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Beast}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object: game.EventPermanentReference(),
											PowerDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountBlockingCreatures,
												Multiplier: 1,
												Object:     game.EventPermanentReference(),
											}),
											ToughnessDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountBlockingCreatures,
												Multiplier: 1,
												Object:     game.EventPermanentReference(),
											}),
											Duration: game.DurationUntilEndOfTurn,
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
			Whenever a Beast becomes blocked, it gets +1/+1 until end of turn for each creature blocking it.
		`,
		},
	}
}
