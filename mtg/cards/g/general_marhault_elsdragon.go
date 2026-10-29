package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GeneralMarhaultElsdragon is the card definition for General Marhault Elsdragon.
//
// Type: Legendary Creature — Elf Warrior
// Cost: {2}{R}{G}
//
// Oracle text:
//
//	Whenever a creature you control becomes blocked, it gets +3/+3 until end of turn for each creature blocking it.
var GeneralMarhaultElsdragon = newGeneralMarhaultElsdragon

func newGeneralMarhaultElsdragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "General Marhault Elsdragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Warrior},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a creature you control becomes blocked, it gets +3/+3 until end of turn for each creature blocking it.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerBecameBlocked,
							BlockedScope:     game.BlockedTriggerOnce,
							Controller:       game.TriggerControllerYou,
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
										Primitive: game.ModifyPT{
											Object: game.EventPermanentReference(),
											PowerDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountBlockingCreatures,
												Multiplier: 3,
												Object:     game.EventPermanentReference(),
											}),
											ToughnessDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountBlockingCreatures,
												Multiplier: 3,
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
			Whenever a creature you control becomes blocked, it gets +3/+3 until end of turn for each creature blocking it.
		`,
		},
	}
}
