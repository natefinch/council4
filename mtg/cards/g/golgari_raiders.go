package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GolgariRaiders is the card definition for Golgari Raiders.
//
// Type: Creature — Elf Warrior
// Cost: {3}{G}
//
// Oracle text:
//
//	Haste
//	Undergrowth — This creature enters with a +1/+1 counter on it for each creature card in your graveyard.
var GolgariRaiders = newGolgariRaiders

func newGolgariRaiders() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Golgari Raiders",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("Undergrowth — This creature enters with a +1/+1 counter on it for each creature card in your graveyard.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountCountCardsInZone,
					Multiplier: 1,
					Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
					CardZone:   zone.Graveyard,
					Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
				})}),
			},
			OracleText: `
			Haste
			Undergrowth — This creature enters with a +1/+1 counter on it for each creature card in your graveyard.
		`,
		},
	}
}
