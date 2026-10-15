package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RenariMerchantOfMarvels is the card definition for Renari, Merchant of Marvels.
//
// Type: Legendary Creature — Dragon Artificer
// Cost: {3}{U}
//
// Oracle text:
//
//	You may cast Dragon spells and artifact spells as though they had flash.
//	Choose a Background (You can have a Background as a second commander.)
var RenariMerchantOfMarvels = newRenariMerchantOfMarvels

func newRenariMerchantOfMarvels() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Renari, Merchant of Marvels",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dragon, types.Artificer},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "You may cast Dragon spells and artifact spells as though they had flash.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsAsThoughFlash,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellCharacteristicFilters: []game.CharacteristicFilter{
								game.CharacteristicFilter{
									Subtypes: []types.Sub{types.Dragon},
								},
								game.CharacteristicFilter{
									Types: []types.Card{types.Artifact},
								},
							},
						},
					},
				},
				game.ChooseABackgroundStaticBody,
			},
			OracleText: `
			You may cast Dragon spells and artifact spells as though they had flash.
			Choose a Background (You can have a Background as a second commander.)
		`,
		},
	}
}
