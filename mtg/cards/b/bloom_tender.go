package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloomTender is the card definition for Bloom Tender.
//
// Type: Creature — Elf Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	Vivid — {T}: For each color among permanents you control, add one mana of that color.
var BloomTender = newBloomTender

func newBloomTender() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bloom Tender",
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
				game.TapManaEachControlledColorAbility("Vivid — {T}: For each color among permanents you control, add one mana of that color.", game.Selection{Controller: game.ControllerYou}),
			},
			OracleText: `
			Vivid — {T}: For each color among permanents you control, add one mana of that color.
		`,
		},
	}
}
