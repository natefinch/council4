package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArchonOfAbsolution is the card definition for Archon of Absolution.
//
// Type: Creature — Archon
// Cost: {3}{W}
//
// Oracle text:
//
//	Flying
//	Protection from white (This creature can't be blocked, targeted, dealt damage, enchanted, or equipped by anything white.)
//	Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.
var ArchonOfAbsolution = newArchonOfAbsolution

func newArchonOfAbsolution() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Archon of Absolution",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Archon},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ProtectionFromColorsStaticAbility(color.White),
				game.StaticAbility{
					Text: "Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                           game.RuleEffectAttackTaxPerCreature,
							Duration:                       game.DurationPermanent,
							AffectedPlayer:                 game.PlayerYou,
							AttackTaxGeneric:               1,
							AttackTaxIncludesPlaneswalkers: true,
						},
					},
				},
			},
			OracleText: `
			Flying
			Protection from white (This creature can't be blocked, targeted, dealt damage, enchanted, or equipped by anything white.)
			Creatures can't attack you or planeswalkers you control unless their controller pays {1} for each of those creatures.
		`,
		},
	}
}
