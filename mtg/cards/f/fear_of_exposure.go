package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FearOfExposure is the card definition for Fear of Exposure.
//
// Type: Enchantment Creature — Nightmare
// Cost: {2}{G}
//
// Oracle text:
//
//	As an additional cost to cast this spell, tap two untapped creatures and/or lands you control.
//	Trample
var FearOfExposure = newFearOfExposure

func newFearOfExposure() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Fear of Exposure",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Nightmare},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalTapPermanents,
					Text:               "tap two untapped creatures and/or lands you control",
					Amount:             2,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
					PermanentTypeAlt:   types.Land,
				},
			},
			OracleText: `
			As an additional cost to cast this spell, tap two untapped creatures and/or lands you control.
			Trample
		`,
		},
	}
}
