package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LoreholdExcavation is the card definition for Lorehold Excavation.
//
// Type: Enchantment
// Cost: {R}{W}
//
// Oracle text:
//
//	At the beginning of your end step, mill a card. If a land card was milled this way, you gain 1 life. Otherwise, this enchantment deals 1 damage to each opponent. (To mill a card, put the top card of your library into your graveyard.)
//	{5}, Exile a creature card from your graveyard: Create a tapped 3/2 red and white Spirit creature token.
var LoreholdExcavation = newLoreholdExcavation

func newLoreholdExcavation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Lorehold Excavation",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}, Exile a creature card from your graveyard: Create a tapped 3/2 red and white Spirit creature token.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:          cost.AdditionalExile,
							Text:          "Exile a creature card from your graveyard",
							Amount:        1,
							MatchCardType: true,
							CardType:      types.Creature,
							Source:        zone.Graveyard,
							CounterKind:   counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:      game.Fixed(1),
											Source:      game.TokenDef(loreholdExcavationToken),
											EntryTapped: true,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your end step, mill a card. If a land card was milled this way, you gain 1 life. Otherwise, this enchantment deals 1 damage to each opponent. (To mill a card, put the top card of your library into your graveyard.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepEnd,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:      game.Fixed(1),
											Player:      game.ControllerReference(),
											Destination: zone.Graveyard,
										},
										PublishResult: game.ResultKey("if-you-do"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("if-you-do")},
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:             game.ResultKey("if-you-do"),
											Succeeded:       game.TriTrue,
											ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
											CardOnly:        true,
										}),
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(1),
											Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:             game.ResultKey("if-you-do"),
											Succeeded:       game.TriTrue,
											ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
											CardOnly:        true,
											Negate:          true,
										}),
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
			At the beginning of your end step, mill a card. If a land card was milled this way, you gain 1 life. Otherwise, this enchantment deals 1 damage to each opponent. (To mill a card, put the top card of your library into your graveyard.)
			{5}, Exile a creature card from your graveyard: Create a tapped 3/2 red and white Spirit creature token.
		`,
		},
	}
}

var loreholdExcavationToken = newLoreholdExcavationToken()

func newLoreholdExcavationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
