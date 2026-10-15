package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragonSpirit is the card definition for Dragon Spirit.
//
// Type: Token Creature — Dragon Spirit
//
// Oracle text:
//   When this creature deals damage, sacrifice it.

// DragonSpiritToken5ec570079912404f948ecd174e2e8331 is the card definition for Dragon Spirit.
var DragonSpiritToken5ec570079912404f948ecd174e2e8331 = newDragonSpiritToken5ec570079912404f948ecd174e2e8331()

func newDragonSpiritToken5ec570079912404f948ecd174e2e8331() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:      "Dragon Spirit",
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon, types.Spirit},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature deals damage, sacrifice it.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventDamageDealt,
							Source:          game.TriggerSourceSelf,
							Subject:         game.TriggerSubjectDamageSource,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Sacrifice{
											Object: game.EventPermanentReference(),
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
			When this creature deals damage, sacrifice it.
		`,
		},
	}
}
