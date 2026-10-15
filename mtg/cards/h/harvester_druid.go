package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HarvesterDruid is the card definition for Harvester Druid.
//
// Type: Creature — Human Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	{T}: Add one mana of any color that a land you control could produce.
var HarvesterDruid = newHarvesterDruid

func newHarvesterDruid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Harvester Druid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Druid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.TapManaLandsProduceAbility(game.PlayerYou, false),
			},
			OracleText: `
			{T}: Add one mana of any color that a land you control could produce.
		`,
		},
	}
}
