package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// StitchedDrake is the card definition for Stitched Drake.
//
// Type: Creature — Zombie Drake
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	As an additional cost to cast this spell, exile a creature card from your graveyard.
//	Flying
var StitchedDrake = newStitchedDrake

func newStitchedDrake() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Stitched Drake",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Drake},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:          cost.AdditionalExile,
					Text:          "exile a creature card from your graveyard",
					Amount:        1,
					Source:        zone.Graveyard,
					MatchCardType: true,
					CardType:      types.Creature,
				},
			},
			OracleText: `
			As an additional cost to cast this spell, exile a creature card from your graveyard.
			Flying
		`,
		},
	}
}
