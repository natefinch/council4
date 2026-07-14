package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UnderworldBreach is the card definition for Underworld Breach.
//
// Type: Enchantment
// Cost: {1}{R}
//
// Oracle text:
//
//	Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus exile three other cards from your graveyard. (You may cast cards from your graveyard for their escape cost.)
//	At the beginning of the end step, sacrifice this enchantment.
var UnderworldBreach = newUnderworldBreach

func newUnderworldBreach() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Underworld Breach",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectGrantGraveyardCardKeyword,
							AffectedPlayer: game.PlayerYou,
							CardSelection:  game.Selection{ExcludedTypes: []types.Card{types.Land}},
							GrantedKeyword: game.Escape,
							GraveyardCastCost: game.GraveyardCastGrantCost{
								UseCardManaCost: true,
								AdditionalCosts: []cost.Additional{
									{
										Kind:          cost.AdditionalExile,
										Text:          "Exile three other cards from your graveyard",
										Amount:        3,
										Source:        zone.Graveyard,
										ExcludeSource: true,
									},
								},
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepEnd,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Sacrifice{
									Object: game.SourceCardPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Each nonland card in your graveyard has escape. The escape cost is equal to the card's mana cost plus exile three other cards from your graveyard. (You may cast cards from your graveyard for their escape cost.)
			At the beginning of the end step, sacrifice this enchantment.
		`,
		},
	}
}
