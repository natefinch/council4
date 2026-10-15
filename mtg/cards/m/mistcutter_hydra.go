package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MistcutterHydra is the card definition for Mistcutter Hydra.
//
// Type: Creature — Hydra
// Cost: {X}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Haste, protection from blue
//	This creature enters with X +1/+1 counters on it.
var MistcutterHydra = newMistcutterHydra

func newMistcutterHydra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Mistcutter Hydra",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hydra},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.HasteStaticBody,
				game.ProtectionFromColorsStaticAbility(color.Blue),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with X +1/+1 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}),
			},
			OracleText: `
			This spell can't be countered.
			Haste, protection from blue
			This creature enters with X +1/+1 counters on it.
		`,
		},
	}
}
