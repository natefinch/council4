package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WreckingCrew is the card definition for Wrecking Crew.
//
// Type: Creature — Human Warrior
// Cost: {4}{R}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
var WreckingCrew = newWreckingCrew

func newWreckingCrew() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Wrecking Crew",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
		`,
		},
	}
}
