package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WizardsOfThay is the card definition for Wizards of Thay.
//
// Type: Creature — Human Wizard
// Cost: {3}{U}
//
// Oracle text:
//
//	Myriad (Whenever this creature attacks, for each opponent other than defending player, you may create a token copy that's tapped and attacking that player or a planeswalker they control. Exile the tokens at end of combat.)
//	Instant and sorcery spells you cast cost {1} less to cast.
//	You may cast sorcery spells as though they had flash.
var WizardsOfThay = newWizardsOfThay

func newWizardsOfThay() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Wizards of Thay",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Instant}},
								GenericReduction: 1,
							},
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Sorcery}},
								GenericReduction: 1,
							},
						},
					},
				},
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsAsThoughFlash,
							AffectedPlayer: game.PlayerYou,
							SpellTypes:     []types.Card{types.Sorcery},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.MyriadTriggeredBody,
			},
			OracleText: `
			Myriad (Whenever this creature attacks, for each opponent other than defending player, you may create a token copy that's tapped and attacking that player or a planeswalker they control. Exile the tokens at end of combat.)
			Instant and sorcery spells you cast cost {1} less to cast.
			You may cast sorcery spells as though they had flash.
		`,
		},
	}
}
