package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CurseOfVitality is the card definition for Curse of Vitality.
//
// Type: Enchantment — Aura Curse
// Cost: {2}{W}
//
// Oracle text:
//
//	Enchant player
//	Whenever enchanted player is attacked, you gain 2 life. Each opponent attacking that player does the same.
var CurseOfVitality = newCurseOfVitality

func newCurseOfVitality() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Curse of Vitality",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
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
								Primitive: game.GainLife{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.GainLife{
									Amount:      game.Fixed(2),
									PlayerGroup: game.OpponentsAttackingTriggerPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant player
			Whenever enchanted player is attacked, you gain 2 life. Each opponent attacking that player does the same.
		`,
		},
	}
}
