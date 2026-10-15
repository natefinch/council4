package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WrithingNecromass is the card definition for Writhing Necromass.
//
// Type: Creature — Zombie Giant
// Cost: {6}{B}
//
// Oracle text:
//
//	This spell costs {1} less to cast for each creature card in your graveyard.
//	Deathtouch
var WrithingNecromass = newWrithingNecromass

func newWrithingNecromass() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Writhing Necromass",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Giant},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
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
				game.DeathtouchStaticBody,
			},
			OracleText: `
			This spell costs {1} less to cast for each creature card in your graveyard.
			Deathtouch
		`,
		},
	}
}
