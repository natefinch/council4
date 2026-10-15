package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CurseOfTheBloodyTome is the card definition for Curse of the Bloody Tome.
//
// Type: Enchantment — Aura Curse
// Cost: {2}{U}
//
// Oracle text:
//
//	Enchant player
//	At the beginning of enchanted player's upkeep, that player mills two cards.
var CurseOfTheBloodyTome = newCurseOfTheBloodyTome

func newCurseOfTheBloodyTome() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Curse of the Bloody Tome",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura, types.Curse},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "player",
					Allow:      game.TargetAllowPlayer,
				}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of enchanted player's upkeep, that player mills two cards.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:                             game.EventBeginningOfStep,
							StackObjectKind:                   game.StackSpell,
							CounterKind:                       counter.PlusOnePlusOne,
							Step:                              game.StepUpkeep,
							StepPlayerIsSourceEnchantedPlayer: true,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:      game.Fixed(2),
											Player:      game.EventPlayerReference(),
											Destination: zone.Graveyard,
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
			Enchant player
			At the beginning of enchanted player's upkeep, that player mills two cards.
		`,
		},
	}
}
