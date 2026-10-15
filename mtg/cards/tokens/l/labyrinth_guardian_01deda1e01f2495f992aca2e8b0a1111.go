package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LabyrinthGuardian is the card definition for Labyrinth Guardian.
//
// Type: Token Creature — Zombie Illusion Warrior
//
// Oracle text:
//   When Labyrinth Guardian becomes the target of a spell, sacrifice it.

// LabyrinthGuardianToken01deda1e01f2495f992aca2e8b0a1111 is the card definition for Labyrinth Guardian.
var LabyrinthGuardianToken01deda1e01f2495f992aca2e8b0a1111 = newLabyrinthGuardianToken01deda1e01f2495f992aca2e8b0a1111()

func newLabyrinthGuardianToken01deda1e01f2495f992aca2e8b0a1111() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Labyrinth Guardian",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Illusion, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Labyrinth Guardian becomes the target of a spell, sacrifice it.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:                game.EventObjectBecameTarget,
							Source:               game.TriggerSourceSelf,
							MatchStackObjectKind: true,
							StackObjectKind:      game.StackSpell,
							CounterKind:          counter.PlusOnePlusOne,
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
			When Labyrinth Guardian becomes the target of a spell, sacrifice it.
		`,
		},
	}
}
