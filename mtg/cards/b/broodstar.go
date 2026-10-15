package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Broodstar is the card definition for Broodstar.
//
// Type: Creature — Beast
// Cost: {8}{U}{U}
//
// Oracle text:
//
//	Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
//	Flying
//	Broodstar's power and toughness are each equal to the number of artifacts you control.
var Broodstar = newBroodstar

func newBroodstar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Broodstar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
				cost.U,
				cost.U,
			}),
			Colors:           []color.Color{color.Blue},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Beast},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerArtifactCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerArtifactCount}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This spell costs {1} less to cast for each artifact you control.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou},
							},
						},
					},
				},
				game.FlyingStaticBody,
			},
			OracleText: `
			Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
			Flying
			Broodstar's power and toughness are each equal to the number of artifacts you control.
		`,
		},
	}
}
