package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CrypticSerpent is the card definition for Cryptic Serpent.
//
// Type: Creature — Serpent
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	This spell costs {1} less to cast for each instant and sorcery card in your graveyard.
var CrypticSerpent = newCrypticSerpent

func newCrypticSerpent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Cryptic Serpent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Serpent},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each instant and sorcery card in your graveyard.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
								CountZone:          opt.Val(zone.Graveyard),
							},
						},
					},
				},
			},
			OracleText: `
			This spell costs {1} less to cast for each instant and sorcery card in your graveyard.
		`,
		},
	}
}
