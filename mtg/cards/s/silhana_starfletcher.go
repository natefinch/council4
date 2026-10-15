package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SilhanaStarfletcher is the card definition for Silhana Starfletcher.
//
// Type: Creature — Elf Druid Archer
// Cost: {2}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	As this creature enters, choose a color.
//	{T}: Add one mana of the chosen color.
var SilhanaStarfletcher = newSilhanaStarfletcher

func newSilhanaStarfletcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Silhana Starfletcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Druid, types.Archer},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color."),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this creature enters, choose a color."),
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			As this creature enters, choose a color.
			{T}: Add one mana of the chosen color.
		`,
		},
	}
}
