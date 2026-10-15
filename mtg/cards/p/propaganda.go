package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Propaganda is the card definition for Propaganda.
//
// Type: Enchantment
// Cost: {2}{U}
//
// Oracle text:
//
//	Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.
var Propaganda = newPropaganda

func newPropaganda() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Propaganda",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectAttackTax,
							Duration:         game.DurationPermanent,
							AffectedPlayer:   game.PlayerYou,
							AttackTaxGeneric: 2,
						},
					},
				},
			},
			OracleText: `
			Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.
		`,
		},
	}
}
