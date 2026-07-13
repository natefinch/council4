package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RustShieldRampager is the card definition for RustShieldRampager.
//
// Type: Creature — Raccoon Warrior
// Cost: {3}{G}
//
// Oracle text:
//
//	Offspring {2} (You may pay an additional {2} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
//	This creature can't be blocked by creatures with power 2 or less.
var RustShieldRampager = newRustShieldRampager

func newRustShieldRampager() *game.CardDef {
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
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.OffspringStaticAbility(cost.Mana{cost.O(2)}),
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerLessOrEqual,
								Power: 2,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.OffspringEnterTriggeredAbility(),
			},
			OracleText: `
			Offspring {2} (You may pay an additional {2} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
			This creature can't be blocked by creatures with power 2 or less.
		`,
		},
	}
}
