package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuirionExplorer is the card definition for Quirion Explorer.
//
// Type: Creature — Elf Druid Scout
// Cost: {1}{G}
//
// Oracle text:
//
//	{T}: Add one mana of any color that a land an opponent controls could produce.
var QuirionExplorer = newQuirionExplorer

func newQuirionExplorer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Quirion Explorer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Druid, types.Scout},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.TapManaLandsProduceAbility(game.PlayerOpponent, false),
			},
			OracleText: `
			{T}: Add one mana of any color that a land an opponent controls could produce.
		`,
		},
	}
}
