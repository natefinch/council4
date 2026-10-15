package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AnglerTurtle is the card definition for Angler Turtle.
//
// Type: Creature — Turtle
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Hexproof
//	Creatures your opponents control attack each combat if able.
var AnglerTurtle = newAnglerTurtle

func newAnglerTurtle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Angler Turtle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Turtle},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
				game.StaticAbility{
					Text: "Creatures your opponents control attack each combat if able.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectMustAttack,
							Duration:          game.DurationPermanent,
							PermanentTypes:    []types.Card{types.Creature},
							AffectedSelection: game.Selection{Controller: game.ControllerOpponent},
						},
					},
				},
			},
			OracleText: `
			Hexproof
			Creatures your opponents control attack each combat if able.
		`,
		},
	}
}
