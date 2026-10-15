package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HazeOfPollen is the card definition for Haze of Pollen.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Prevent all combat damage that would be dealt this turn.
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var HazeOfPollen = newHazeOfPollen

func newHazeOfPollen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Haze of Pollen",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.PreventDamage{
							All:        true,
							CombatOnly: true,
							Global:     true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Prevent all combat damage that would be dealt this turn.
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
