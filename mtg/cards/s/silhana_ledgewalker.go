package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SilhanaLedgewalker is the card definition for Silhana Ledgewalker.
//
// Type: Creature — Elf Rogue
// Cost: {1}{G}
//
// Oracle text:
//
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
//	This creature can't be blocked except by creatures with flying.
var SilhanaLedgewalker = newSilhanaLedgewalker

func newSilhanaLedgewalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Silhana Ledgewalker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedExceptBy,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind: game.BlockerRestrictionFlying,
							},
						},
					},
				},
			},
			OracleText: `
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
			This creature can't be blocked except by creatures with flying.
		`,
		},
	}
}
