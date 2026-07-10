package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SavageVentmaw is the card definition for Savage Ventmaw.
//
// Type: Creature — Dragon
// Cost: {4}{R}{G}
//
// Oracle text:
//
//	Flying
//	Whenever this creature attacks, add {R}{R}{R}{G}{G}{G}. Until end of turn, you don't lose this mana as steps and phases end.
var SavageVentmaw = newSavageVentmaw

func newSavageVentmaw() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Savage Ventmaw",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
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
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.R,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.R,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.R,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.G,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.G,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.G,
									PersistUntilEndOfTurn: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever this creature attacks, add {R}{R}{R}{G}{G}{G}. Until end of turn, you don't lose this mana as steps and phases end.
		`,
		},
	}
}
