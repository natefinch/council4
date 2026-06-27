package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IronLadYoungAvenger is the card definition for Iron Lad, Young Avenger.
//
// Type: Legendary Artifact Creature — Human Hero
// Cost: {2}{U/R}
//
// Oracle text:
//
//	Flying (This creature can't be blocked except by creatures with flying or reach.)
//	Noncreature spells you cast cost {1} less to cast.
var IronLadYoungAvenger = newIronLadYoungAvenger()

func newIronLadYoungAvenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Iron Lad, Young Avenger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.U, mana.R),
			}),
			Colors:     []color.Color{color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Hero},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{ExcludedTypes: []types.Card{types.Creature}},
								GenericReduction: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Flying (This creature can't be blocked except by creatures with flying or reach.)
			Noncreature spells you cast cost {1} less to cast.
		`,
		},
	}
}
