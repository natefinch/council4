package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Pirate is the card definition for Pirate.
//
// Type: Token Creature — Pirate
//
// Oracle text:
//   This creature can't block.
//   Creatures you control attack each combat if able.

// PirateTokenf24f68d06e814294a0848ec8f36fef18 is the card definition for Pirate.
var PirateTokenf24f68d06e814294a0848ec8f36fef18 = newPirateTokenf24f68d06e814294a0848ec8f36fef18()

func newPirateTokenf24f68d06e814294a0848ec8f36fef18() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:      "Pirate",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Pirate},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
				game.StaticAbility{
					Text: "Creatures you control attack each combat if able.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectMustAttack,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
						},
					},
				},
			},
			OracleText: `
			This creature can't block.
			Creatures you control attack each combat if able.
		`,
		},
	}
}
