package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DreadWarlock is the card definition for Dread Warlock.
//
// Type: Creature — Human Wizard Warlock
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	This creature can't be blocked except by black creatures.
var DreadWarlock = newDreadWarlock()

func newDreadWarlock() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dread Warlock",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard, types.Warlock},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedExceptBy,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionColor,
								Color: color.Black,
							},
						},
					},
				},
			},
			OracleText: `
			This creature can't be blocked except by black creatures.
		`,
		},
	}
}
