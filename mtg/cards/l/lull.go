package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Lull is the card definition for Lull.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Prevent all combat damage that would be dealt this turn.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var Lull = newLull

func newLull() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lull",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
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
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
