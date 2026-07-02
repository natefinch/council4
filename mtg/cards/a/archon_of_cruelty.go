package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArchonOfCruelty is the card definition for Archon of Cruelty.
//
// Type: Creature — Archon
// Cost: {6}{B}{B}
//
// Oracle text:
//
//	Flying
//	Whenever this creature enters or attacks, target opponent sacrifices a creature or planeswalker of their choice, discards a card, and loses 3 life. You draw a card and gain 3 life.
var ArchonOfCruelty = newArchonOfCruelty()

func newArchonOfCruelty() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Archon of Cruelty",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Archon},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventPermanentEnteredBattlefield,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventAttackerDeclared,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.TargetPlayerReference(0),
									Selection: game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}},
								},
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.TargetPlayerReference(0),
								},
							},
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(3),
									Player: game.TargetPlayerReference(0),
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever this creature enters or attacks, target opponent sacrifices a creature or planeswalker of their choice, discards a card, and loses 3 life. You draw a card and gain 3 life.
		`,
		},
	}
}
