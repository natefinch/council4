package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HolyDay is the card definition for Holy Day.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Prevent all combat damage that would be dealt this turn.
var HolyDay = newHolyDay

func newHolyDay() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Holy Day",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
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
		`,
		},
	}
}
