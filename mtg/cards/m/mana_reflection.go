package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ManaReflection is the card definition for Mana Reflection.
//
// Type: Enchantment
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	If you tap a permanent for mana, it produces twice as much of that mana instead.
var ManaReflection = newManaReflection()

func newManaReflection() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Mana Reflection",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                     game.RuleEffectManaProductionMultiplier,
							ManaProductionMultiplier: 2,
						},
					},
				},
			},
			OracleText: `
			If you tap a permanent for mana, it produces twice as much of that mana instead.
		`,
		},
	}
}
