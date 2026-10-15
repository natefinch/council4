package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Ghoultree is the card definition for Ghoultree.
//
// Type: Creature — Zombie Treefolk
// Cost: {7}{G}
//
// Oracle text:
//
//	This spell costs {1} less to cast for each creature card in your graveyard.
var Ghoultree = newGhoultree

func newGhoultree() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ghoultree",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Treefolk},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 10}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each creature card in your graveyard.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Creature}},
								CountZone:          opt.Val(zone.Graveyard),
							},
						},
					},
				},
			},
			OracleText: `
			This spell costs {1} less to cast for each creature card in your graveyard.
		`,
		},
	}
}
