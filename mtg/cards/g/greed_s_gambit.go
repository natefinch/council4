package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GreedSGambit is the card definition for Greed's Gambit.
//
// Type: Enchantment
// Cost: {3}{B}
//
// Oracle text:
//
//	When this enchantment enters, you draw three cards, gain 6 life, and create three 2/1 black Bat creature tokens with flying.
//	At the beginning of your end step, you discard a card, lose 2 life, and sacrifice a creature.
//	When this enchantment leaves the battlefield, you discard three cards, lose 6 life, and sacrifice three creatures.
var GreedSGambit = newGreedSGambit()

func newGreedSGambit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Greed's Gambit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
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
								Primitive: game.Draw{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(6),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(3),
									Source: game.TokenDef(greedSGambitToken),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.ControllerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(6),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(3),
									Player:    game.ControllerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this enchantment enters, you draw three cards, gain 6 life, and create three 2/1 black Bat creature tokens with flying.
			At the beginning of your end step, you discard a card, lose 2 life, and sacrifice a creature.
			When this enchantment leaves the battlefield, you discard three cards, lose 6 life, and sacrifice three creatures.
		`,
		},
	}
}

var greedSGambitToken = newGreedSGambitToken()

func newGreedSGambitToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bat",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bat},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
