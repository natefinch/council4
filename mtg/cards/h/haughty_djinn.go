package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HaughtyDjinn is the card definition for Haughty Djinn.
//
// Type: Creature — Djinn
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	Flying
//	Haughty Djinn's power is equal to the number of instant and sorcery cards in your graveyard.
//	Instant and sorcery spells you cast cost {1} less to cast.
var HaughtyDjinn = newHaughtyDjinn

func newHaughtyDjinn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Haughty Djinn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors:       []color.Color{color.Blue},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Djinn},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerInstantOrSorceryCardsInGraveyard}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
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
			Flying
			Haughty Djinn's power is equal to the number of instant and sorcery cards in your graveyard.
			Instant and sorcery spells you cast cost {1} less to cast.
		`,
		},
	}
}
