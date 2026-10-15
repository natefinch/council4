package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FootEnforcer is the card definition for Foot Enforcer.
//
// Type: Creature — Ninja
//
// Oracle text:
//   This creature can't block.
//   This creature can't be blocked by more than one creature.

// FootEnforcerToken7124177655ae4e3cbf013387a9fca88c is the card definition for Foot Enforcer.
var FootEnforcerToken7124177655ae4e3cbf013387a9fca88c = newFootEnforcerToken7124177655ae4e3cbf013387a9fca88c()

func newFootEnforcerToken7124177655ae4e3cbf013387a9fca88c() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Foot Enforcer",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ninja},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
				game.StaticAbility{
					Text: "This creature can't be blocked by more than one creature.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByMoreThanOne,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			This creature can't block.
			This creature can't be blocked by more than one creature.
		`,
		},
	}
}
