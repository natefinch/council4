package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AbstractPaintmage is the card definition for Abstract Paintmage.
//
// Type: Creature — Djinn Sorcerer
// Cost: {U}{U/R}{R}
//
// Oracle text:
//
//	At the beginning of your first main phase, add {U}{R}. Spend this mana only to cast instant and sorcery spells.
var AbstractPaintmage = newAbstractPaintmage()

func newAbstractPaintmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Abstract Paintmage",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.HybridMana(mana.U, mana.R),
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Djinn, types.Sorcerer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepPrecombatMain,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.U,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastInstantOrSorcerySpell,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.R,
									SpendRider: opt.Val(game.ManaSpendRider{
										Condition:   game.ManaSpendCastInstantOrSorcerySpell,
										Restriction: game.ManaSpendRestrictedToCondition,
									}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your first main phase, add {U}{R}. Spend this mana only to cast instant and sorcery spells.
		`,
		},
	}
}
