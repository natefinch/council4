package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EelUmbra is the card definition for Eel Umbra.
//
// Type: Enchantment — Aura
// Cost: {1}{U}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Enchant creature
//	Enchanted creature gets +1/+1.
//	Umbra armor (If enchanted creature would be destroyed, instead remove all damage from it and destroy this Aura.)
var EelUmbra = newEelUmbra

func newEelUmbra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Eel Umbra",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature gets +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
				game.UmbraArmorStaticBody,
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			Enchant creature
			Enchanted creature gets +1/+1.
			Umbra armor (If enchanted creature would be destroyed, instead remove all damage from it and destroy this Aura.)
		`,
		},
	}
}
