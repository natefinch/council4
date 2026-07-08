package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArchdemonOfUnx is the card definition for Archdemon of Unx.
//
// Type: Creature — Demon
// Cost: {5}{B}{B}
//
// Oracle text:
//
//	Flying, trample
//	At the beginning of your upkeep, sacrifice a non-Zombie creature, then create a 2/2 black Zombie creature token.
var ArchdemonOfUnx = newArchdemonOfUnx

func newArchdemonOfUnx() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Archdemon of Unx",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Demon},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.ControllerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Zombie")},
								},
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(archdemonOfUnxToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, trample
			At the beginning of your upkeep, sacrifice a non-Zombie creature, then create a 2/2 black Zombie creature token.
		`,
		},
	}
}

var archdemonOfUnxToken = newArchdemonOfUnxToken()

func newArchdemonOfUnxToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
