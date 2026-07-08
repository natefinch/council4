package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpectacleMage is the card definition for Spectacle Mage.
//
// Type: Creature — Bird Shaman
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	Flying
//	Instant and sorcery spells you cast with mana value 5 or greater cost {1} less to cast.
var SpectacleMage = newSpectacleMage

func newSpectacleMage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Spectacle Mage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Instant}, ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5})},
								GenericReduction: 1,
							},
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Sorcery}, ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 5})},
								GenericReduction: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Flying
			Instant and sorcery spells you cast with mana value 5 or greater cost {1} less to cast.
		`,
		},
	}
}
