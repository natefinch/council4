package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OreScaleGuardian is the card definition for Ore-Scale Guardian.
//
// Type: Creature — Dragon
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	This spell costs {1} less to cast for each land card in your graveyard.
//	Flying, haste
var OreScaleGuardian = newOreScaleGuardian

func newOreScaleGuardian() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ore-Scale Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each land card in your graveyard.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Land}},
								CountZone:          opt.Val(zone.Graveyard),
							},
						},
					},
				},
				game.FlyingStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			This spell costs {1} less to cast for each land card in your graveyard.
			Flying, haste
		`,
		},
	}
}
