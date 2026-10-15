package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InkfathomInfiltrator is the card definition for Inkfathom Infiltrator.
//
// Type: Creature — Merfolk Rogue
// Cost: {U/B}{U/B}
//
// Oracle text:
//
//	This creature can't block and can't be blocked.
var InkfathomInfiltrator = newInkfathomInfiltrator

func newInkfathomInfiltrator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Inkfathom Infiltrator",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.U, mana.B),
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't block and can't be blocked.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlocked,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			This creature can't block and can't be blocked.
		`,
		},
	}
}
