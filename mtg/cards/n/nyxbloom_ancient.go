package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NyxbloomAncient is the card definition for Nyxbloom Ancient.
//
// Type: Enchantment Creature — Elemental
// Cost: {4}{G}{G}{G}
//
// Oracle text:
//
//	Trample
//	If you tap a permanent for mana, it produces three times as much of that mana instead.
var NyxbloomAncient = newNyxbloomAncient

func newNyxbloomAncient() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nyxbloom Ancient",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                     game.RuleEffectManaProductionMultiplier,
							ManaProductionMultiplier: 3,
						},
					},
				},
			},
			OracleText: `
			Trample
			If you tap a permanent for mana, it produces three times as much of that mana instead.
		`,
		},
	}
}
