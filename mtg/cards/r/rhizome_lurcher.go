package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RhizomeLurcher is the card definition for Rhizome Lurcher.
//
// Type: Creature — Fungus Zombie
// Cost: {2}{B}{G}
//
// Oracle text:
//
//	Undergrowth — This creature enters with a number of +1/+1 counters on it equal to the number of creature cards in your graveyard.
var RhizomeLurcher = newRhizomeLurcher()

func newRhizomeLurcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Rhizome Lurcher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.G,
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fungus, types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("Undergrowth — This creature enters with a number of +1/+1 counters on it equal to the number of creature cards in your graveyard.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Dynamic: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountCountCardsInZone,
					Multiplier: 1,
					Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
					CardZone:   zone.Graveyard,
					Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
				})}),
			},
			OracleText: `
			Undergrowth — This creature enters with a number of +1/+1 counters on it equal to the number of creature cards in your graveyard.
		`,
		},
	}
}
