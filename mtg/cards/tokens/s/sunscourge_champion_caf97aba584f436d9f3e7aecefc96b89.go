package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SunscourgeChampion is the card definition for Sunscourge Champion.
//
// Type: Token Creature — Zombie Human Wizard
//
// Oracle text:
//   When Sunscourge Champion enters the battlefield, you gain life equal to its power.

// SunscourgeChampionTokencaf97aba584f436d9f3e7aecefc96b89 is the card definition for Sunscourge Champion.
var SunscourgeChampionTokencaf97aba584f436d9f3e7aecefc96b89 = newSunscourgeChampionTokencaf97aba584f436d9f3e7aecefc96b89()

func newSunscourgeChampionTokencaf97aba584f436d9f3e7aecefc96b89() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Sunscourge Champion",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Sunscourge Champion enters the battlefield, you gain life equal to its power.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
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
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountObjectPower,
												Multiplier: 1,
												Object:     game.EventPermanentReference(),
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
			OracleText: `
			When Sunscourge Champion enters the battlefield, you gain life equal to its power.
		`,
		},
	}
}
