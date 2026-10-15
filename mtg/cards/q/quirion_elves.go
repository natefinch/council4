package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuirionElves is the card definition for Quirion Elves.
//
// Type: Creature — Elf Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	As this creature enters, choose a color.
//	{T}: Add {G}.
//	{T}: Add one mana of the chosen color.
var QuirionElves = newQuirionElves

func newQuirionElves() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Quirion Elves",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Druid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
				game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color."),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this creature enters, choose a color."),
			},
			OracleText: `
			As this creature enters, choose a color.
			{T}: Add {G}.
			{T}: Add one mana of the chosen color.
		`,
		},
	}
}
