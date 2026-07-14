package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CurseOfVerbosity is the card definition for Curse of Verbosity.
//
// Type: Enchantment — Aura Curse
// Cost: {2}{U}
//
// Oracle text:
//
//	Enchant player
//	Whenever enchanted player is attacked, you draw a card. Each opponent attacking that player does the same.
var CurseOfVerbosity = newCurseOfVerbosity

func newCurseOfVerbosity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Curse of Verbosity",
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
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                                 game.EventAttackerDeclared,
							OneOrMore:                             true,
							AttackedPlayerIsSourceEnchantedPlayer: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Draw{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsAttackingTriggerPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant player
			Whenever enchanted player is attacked, you draw a card. Each opponent attacking that player does the same.
		`,
		},
	}
}
