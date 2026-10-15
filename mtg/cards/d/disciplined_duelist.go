package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DisciplinedDuelist is the card definition for Disciplined Duelist.
//
// Type: Creature — Human Citizen
// Cost: {G}{W}{U}
//
// Oracle text:
//
//	Double strike
//	This creature enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
var DisciplinedDuelist = newDisciplinedDuelist

func newDisciplinedDuelist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Disciplined Duelist",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Green, color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Citizen},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)", game.CounterPlacement{Kind: counter.Shield, Amount: 1}),
			},
			OracleText: `
			Double strike
			This creature enters with a shield counter on it. (If it would be dealt damage or destroyed, remove a shield counter from it instead.)
		`,
		},
	}
}
