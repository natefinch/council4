package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StinkdrinkerDaredevil is the card definition for Stinkdrinker Daredevil.
//
// Type: Creature — Goblin Rogue
// Cost: {2}{R}
//
// Oracle text:
//
//	Giant spells you cast cost {2} less to cast.
var StinkdrinkerDaredevil = newStinkdrinkerDaredevil

func newStinkdrinkerDaredevil() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Stinkdrinker Daredevil",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Giant spells you cast cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Giant}},
							},
						},
					},
				},
			},
			OracleText: `
			Giant spells you cast cost {2} less to cast.
		`,
		},
	}
}
