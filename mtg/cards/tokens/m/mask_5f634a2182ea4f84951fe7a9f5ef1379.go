package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// Mask is the card definition for Mask.
//
// Type: Token Enchantment — Aura
//
// Oracle text:
//   Enchant permanent
//   Totem armor (If enchanted permanent would be destroyed, instead remove all damage from it and destroy this Aura.)

// MaskToken5f634a2182ea4f84951fe7a9f5ef1379 is the card definition for Mask.
var MaskToken5f634a2182ea4f84951fe7a9f5ef1379 = newMaskToken5f634a2182ea4f84951fe7a9f5ef1379()

func newMaskToken5f634a2182ea4f84951fe7a9f5ef1379() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:     "Mask",
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "permanent",
					Allow:      game.TargetAllowPermanent,
				}),
				game.UmbraArmorStaticBody,
			},
			OracleText: `
			Enchant permanent
			Totem armor (If enchanted permanent would be destroyed, instead remove all damage from it and destroy this Aura.)
		`,
		},
	}
}
