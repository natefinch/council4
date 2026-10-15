package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SalvageSlasher is the card definition for Salvage Slasher.
//
// Type: Artifact Creature — Human Rogue
// Cost: {1}{B}
//
// Oracle text:
//
//	This creature gets +1/+0 for each artifact card in your graveyard.
var SalvageSlasher = newSalvageSlasher

func newSalvageSlasher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Salvage Slasher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature gets +1/+0 for each artifact card in your graveyard.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:        game.DynamicAmountCountCardsInZone,
								Multiplier:  1,
								CounterKind: counter.PlusOnePlusOne,
								Player:      func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:    zone.Graveyard,
								Selection:   &game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							}),
						},
					},
				},
			},
			OracleText: `
			This creature gets +1/+0 for each artifact card in your graveyard.
		`,
		},
	}
}
