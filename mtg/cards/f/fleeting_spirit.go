package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FleetingSpirit is the card definition for Fleeting Spirit.
//
// Type: Creature — Spirit
// Cost: {1}{W}
//
// Oracle text:
//
//	{W}, Exile three cards from your graveyard: This creature gains first strike until end of turn.
//	Discard a card: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.
var FleetingSpirit = newFleetingSpirit

func newFleetingSpirit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Fleeting Spirit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{W}, Exile three cards from your graveyard: This creature gains first strike until end of turn.",
					ManaCost: opt.Val(cost.Mana{cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalExile,
							Text:   "Exile three cards from your graveyard",
							Amount: 3,
							Source: zone.Graveyard,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourceCardPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.FirstStrike,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text: "Discard a card: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalDiscard,
							Text:   "Discard a card",
							Amount: 1,
							Source: zone.Hand,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object:         game.SourcePermanentReference(),
									ExileLinkedKey: game.LinkedKey("delayed-self-blink"),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtBeginningOfNextEndStep,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.PutOnBattlefield{
														Source: game.LinkedBattlefieldSource(game.LinkedKey("delayed-self-blink")),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{W}, Exile three cards from your graveyard: This creature gains first strike until end of turn.
			Discard a card: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.
		`,
		},
	}
}
