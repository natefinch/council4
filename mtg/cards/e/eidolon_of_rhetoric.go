package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EidolonOfRhetoric is the card definition for Eidolon of Rhetoric.
//
// Type: Enchantment Creature — Spirit
// Cost: {2}{W}
//
// Oracle text:
//
//	Each player can't cast more than one spell each turn.
var EidolonOfRhetoric = newEidolonOfRhetoric

func newEidolonOfRhetoric() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Eidolon of Rhetoric",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each player can't cast more than one spell each turn.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCastLimitPerTurn,
							Duration:         game.DurationPermanent,
							CastLimitPerTurn: 1,
						},
					},
				},
			},
			OracleText: `
			Each player can't cast more than one spell each turn.
		`,
		},
	}
}
