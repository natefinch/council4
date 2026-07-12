package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DimirInfiltrator is the card definition for Dimir Infiltrator.
//
// Type: Creature — Spirit
// Cost: {U}{B}
//
// Oracle text:
//
//	This creature can't be blocked.
//	Transmute {1}{U}{B} ({1}{U}{B}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
var DimirInfiltrator = newDimirInfiltrator

func newDimirInfiltrator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Dimir Infiltrator",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeBlockedStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.TransmuteActivatedAbility(cost.Mana{cost.O(1), cost.U, cost.B}, 2),
			},
			OracleText: `
			This creature can't be blocked.
			Transmute {1}{U}{B} ({1}{U}{B}, Discard this card: Search your library for a card with the same mana value as this card, reveal it, put it into your hand, then shuffle. Transmute only as a sorcery.)
		`,
		},
	}
}
