package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RustShieldRampager is the card definition for Rust-Shield Rampager.
//
// Type: Token Creature — Raccoon Warrior
// Cost: {3}{G}
//
// Oracle text:
//   This creature can't be blocked by creatures with power 2 or less.
//   (This token's mana cost is {3}{G}.)

// RustShieldRampagerToken93793ab53c854c7895c2c228e39bd930 is the card definition for Rust-Shield Rampager.
var RustShieldRampagerToken93793ab53c854c7895c2c228e39bd930 = newRustShieldRampagerToken93793ab53c854c7895c2c228e39bd930()

func newRustShieldRampagerToken93793ab53c854c7895c2c228e39bd930() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Rust-Shield Rampager",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Raccoon, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't be blocked by creatures with power 2 or less.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerLessOrEqual,
								Power: 2,
							},
						},
					},
				},
			},
			OracleText: `
			This creature can't be blocked by creatures with power 2 or less.
			(This token's mana cost is {3}{G}.)
		`,
		},
	}
}
