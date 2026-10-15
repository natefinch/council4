package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SyrElenoraTheDiscerning is the card definition for Syr Elenora, the Discerning.
//
// Type: Legendary Creature — Human Knight
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Syr Elenora's power is equal to the number of cards in your hand.
//	When Syr Elenora enters, draw a card.
//	Spells your opponents cast that target Syr Elenora cost {2} more to cast.
var SyrElenoraTheDiscerning = newSyrElenoraTheDiscerning

func newSyrElenoraTheDiscerning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Syr Elenora, the Discerning",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:       []color.Color{color.Blue},
			Supertypes:   []types.Super{types.Legendary},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Human, types.Knight},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Spells your opponents cast that target Syr Elenora cost {2} more to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerOpponent,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								GenericIncrease: 2,
								TargetsSource:   true,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Syr Elenora enters, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Syr Elenora's power is equal to the number of cards in your hand.
			When Syr Elenora enters, draw a card.
			Spells your opponents cast that target Syr Elenora cost {2} more to cast.
		`,
		},
	}
}
