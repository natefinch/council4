package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NagaVitalist is the card definition for Naga Vitalist.
//
// Type: Creature — Snake Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	{T}: Add one mana of any type that a land you control could produce.
var NagaVitalist = newNagaVitalist

func newNagaVitalist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Naga Vitalist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Druid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ManaAbilities: []game.ManaAbility{
				game.TapManaLandsProduceAbility(game.PlayerYou, true),
			},
			OracleText: `
			{T}: Add one mana of any type that a land you control could produce.
		`,
		},
	}
}
