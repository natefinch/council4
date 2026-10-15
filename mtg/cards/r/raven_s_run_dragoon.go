package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RavenSRunDragoon is the card definition for Raven's Run Dragoon.
//
// Type: Creature — Elf Knight
// Cost: {2}{G/W}{G/W}
//
// Oracle text:
//
//	This creature can't be blocked by black creatures.
var RavenSRunDragoon = newRavenSRunDragoon

func newRavenSRunDragoon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Raven's Run Dragoon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Knight},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't be blocked by black creatures.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionColor,
								Color: color.Black,
							},
						},
					},
				},
			},
			OracleText: `
			This creature can't be blocked by black creatures.
		`,
		},
	}
}
