package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RivenTurnbull is the card definition for Riven Turnbull.
//
// Type: Legendary Creature — Human Advisor
// Cost: {5}{U}{B}
//
// Oracle text:
//
//	{T}: Add {B}.
var RivenTurnbull = newRivenTurnbull

func newRivenTurnbull() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Riven Turnbull",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Advisor},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			OracleText: `
			{T}: Add {B}.
		`,
		},
	}
}
