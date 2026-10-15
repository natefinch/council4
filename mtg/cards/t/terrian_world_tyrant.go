package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TerrianWorldTyrant is the card definition for Terrian, World Tyrant.
//
// Type: Legendary Creature — Dinosaur Ooze
// Cost: {2}{G}{G}{G}
//
// Oracle text:
var TerrianWorldTyrant = newTerrianWorldTyrant

func newTerrianWorldTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Terrian, World Tyrant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dinosaur, types.Ooze},
			Power:      opt.Val(game.PT{Value: 9}),
			Toughness:  opt.Val(game.PT{Value: 7}),
		},
	}
}
