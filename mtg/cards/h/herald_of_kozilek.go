package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HeraldOfKozilek is the card definition for Herald of Kozilek.
//
// Type: Creature — Eldrazi Drone
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Colorless spells you cast cost {1} less to cast.
var HeraldOfKozilek = newHeraldOfKozilek

func newHeraldOfKozilek() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Herald of Kozilek",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Drone},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.StaticAbility{
					Text: "Colorless spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{Colorless: true},
							},
						},
					},
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			Colorless spells you cast cost {1} less to cast.
		`,
		},
	}
}
