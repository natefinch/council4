package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpaceMarineDevastator is the card definition for Space Marine Devastator.
//
// Type: Token Creature — Astartes Warrior
//
// Oracle text:
//   When Space Marine Devastator enters the battlefield, destroy up to one target artifact or enchantment.

// SpaceMarineDevastatorTokend74a563883cc480bbb02caecc6a8ff2e is the card definition for Space Marine Devastator.
var SpaceMarineDevastatorTokend74a563883cc480bbb02caecc6a8ff2e = newSpaceMarineDevastatorTokend74a563883cc480bbb02caecc6a8ff2e()

func newSpaceMarineDevastatorTokend74a563883cc480bbb02caecc6a8ff2e() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Space Marine Devastator",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Astartes, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Space Marine Devastator enters the battlefield, destroy up to one target artifact or enchantment.",
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
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 1,
										Constraint: "up to one target artifact or enchantment",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
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
			When Space Marine Devastator enters the battlefield, destroy up to one target artifact or enchantment.
		`,
		},
	}
}
