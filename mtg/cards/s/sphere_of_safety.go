package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphereOfSafety is the card definition for Sphere of Safety.
//
// Type: Enchantment
// Cost: {4}{W}
//
// Oracle text:
//
//	Creatures can't attack you or planeswalkers you control unless their controller pays {X} for each of those creatures, where X is the number of enchantments you control.
var SphereOfSafety = newSphereOfSafety

func newSphereOfSafety() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sphere of Safety",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures can't attack you or planeswalkers you control unless their controller pays {X} for each of those creatures, where X is the number of enchantments you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                           game.RuleEffectAttackTaxPerCreature,
							Duration:                       game.DurationPermanent,
							AffectedPlayer:                 game.PlayerYou,
							CardSelection:                  game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou},
							AttackTaxIncludesPlaneswalkers: true,
						},
					},
				},
			},
			OracleText: `
			Creatures can't attack you or planeswalkers you control unless their controller pays {X} for each of those creatures, where X is the number of enchantments you control.
		`,
		},
	}
}
