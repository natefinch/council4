package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Nulldrifter is the card definition for Nulldrifter.
//
// Type: Creature — Eldrazi Elemental
// Cost: {7}
//
// Oracle text:
//
//	When you cast this spell, draw two cards.
//	Flying
//	Annihilator 1 (Whenever this creature attacks, defending player sacrifices a permanent of their choice.)
//	Evoke {2}{U} (You may cast this spell for its evoke cost. If you do, it's sacrificed when it enters.)
var Nulldrifter = newNulldrifter()

func newNulldrifter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Nulldrifter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Elemental},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:       game.EventSpellCast,
							Source:      game.TriggerSourceSelf,
							Controller:  game.TriggerControllerYou,
							SelfWasCast: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount: game.Fixed(1),
									Player: game.DefendingPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.EvokeSacrificeTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Evoke",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U}),
					Mechanic: cost.AlternativeMechanicEvoke,
				},
			},
			OracleText: `
			When you cast this spell, draw two cards.
			Flying
			Annihilator 1 (Whenever this creature attacks, defending player sacrifices a permanent of their choice.)
			Evoke {2}{U} (You may cast this spell for its evoke cost. If you do, it's sacrificed when it enters.)
		`,
		},
	}
}
