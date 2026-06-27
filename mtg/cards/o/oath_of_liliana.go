package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OathOfLiliana is the card definition for Oath of Liliana.
//
// Type: Legendary Enchantment
// Cost: {2}{B}
//
// Oracle text:
//
//	When Oath of Liliana enters, each opponent sacrifices a creature of their choice.
//	At the beginning of each end step, if a planeswalker entered the battlefield under your control this turn, create a 2/2 black Zombie creature token.
var OathOfLiliana = newOathOfLiliana()

func newOathOfLiliana() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Oath of Liliana",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsReference(),
									Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepEnd,
						},
						InterveningIf: "if a planeswalker entered the battlefield under your control this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:            game.EventPermanentEnteredBattlefield,
								Controller:       game.TriggerControllerYou,
								SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(oathOfLilianaToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When Oath of Liliana enters, each opponent sacrifices a creature of their choice.
			At the beginning of each end step, if a planeswalker entered the battlefield under your control this turn, create a 2/2 black Zombie creature token.
		`,
		},
	}
}

var oathOfLilianaToken = newOathOfLilianaToken()

func newOathOfLilianaToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Zombie",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
