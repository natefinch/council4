package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpellwildOuphe is the card definition for Spellwild Ouphe.
//
// Type: Creature — Ouphe
// Cost: {1}{G}
//
// Oracle text:
//
//	Spells that target this creature cost {2} less to cast.
var SpellwildOuphe = newSpellwildOuphe

func newSpellwildOuphe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Spellwild Ouphe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ouphe},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Spells that target this creature cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:     game.RuleEffectCostModifier,
							Duration: game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								TargetsSource:    true,
							},
						},
					},
				},
			},
			OracleText: `
			Spells that target this creature cost {2} less to cast.
		`,
		},
	}
}
