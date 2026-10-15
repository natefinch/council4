package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinConstruct is the card definition for Goblin Construct.
//
// Type: Token Artifact Creature — Goblin Construct
//
// Oracle text:
//   This creature can't block.
//   At the beginning of your upkeep, this creature deals 1 damage to you.

// GoblinConstructTokenad3edc909f494511aa46fdd0785d8a4c is the card definition for Goblin Construct.
var GoblinConstructTokenad3edc909f494511aa46fdd0785d8a4c = newGoblinConstructTokenad3edc909f494511aa46fdd0785d8a4c()

func newGoblinConstructTokenad3edc909f494511aa46fdd0785d8a4c() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Goblin Construct",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Construct},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, this creature deals 1 damage to you.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
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
			This creature can't block.
			At the beginning of your upkeep, this creature deals 1 damage to you.
		`,
		},
	}
}
