package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinAnarchomancer is the card definition for Goblin Anarchomancer.
//
// Type: Creature — Goblin Shaman
// Cost: {R}{G}
//
// Oracle text:
//
//	Each spell you cast that's red or green costs {1} less to cast.
var GoblinAnarchomancer = newGoblinAnarchomancer

func newGoblinAnarchomancer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Goblin Anarchomancer",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Each spell you cast that's red or green costs {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{ColorsAny: []color.Color{color.Red, color.Green}},
							},
						},
					},
				},
			},
			OracleText: `
			Each spell you cast that's red or green costs {1} less to cast.
		`,
		},
	}
}
