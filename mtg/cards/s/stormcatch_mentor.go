package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StormcatchMentor is the card definition for Stormcatch Mentor.
//
// Type: Creature — Otter Wizard
// Cost: {U}{R}
//
// Oracle text:
//
//	Haste
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	Instant and sorcery spells you cast cost {1} less to cast.
var StormcatchMentor = newStormcatchMentor

func newStormcatchMentor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Stormcatch Mentor",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Otter, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.ProwessStaticBody,
				game.StaticAbility{
					Text: "Instant and sorcery spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Instant}},
							},
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Sorcery}},
							},
						},
					},
				},
			},
			OracleText: `
			Haste
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			Instant and sorcery spells you cast cost {1} less to cast.
		`,
		},
	}
}
