package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TwinflameTravelers is the card definition for Twinflame Travelers.
//
// Type: Creature — Elemental Sorcerer
// Cost: {2}{U}{R}
//
// Oracle text:
//
//	Flying
//	If a triggered ability of another Elemental you control triggers, it triggers an additional time.
var TwinflameTravelers = newTwinflameTravelers()

func newTwinflameTravelers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Twinflame Travelers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Sorcerer},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectAdditionalTriggerForControlledPermanent,
							AffectedSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Elemental")}, ExcludeSource: true},
						},
					},
				},
			},
			OracleText: `
			Flying
			If a triggered ability of another Elemental you control triggers, it triggers an additional time.
		`,
		},
	}
}
