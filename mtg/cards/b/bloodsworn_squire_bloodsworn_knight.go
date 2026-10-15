package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BloodswornSquire is the card definition for Bloodsworn Squire // Bloodsworn Knight.
//
// Type: Creature — Vampire Soldier // Creature — Vampire Knight
// Face: Bloodsworn Knight — Creature — Vampire Knight
//
// Oracle text:
//
//	{1}{B}, Discard a card: This creature gains indestructible until end of turn. Tap it. Then if there are four or more creature cards in your graveyard, transform this creature. (Damage and effects that say "destroy" don't destroy it.)
var BloodswornSquire = newBloodswornSquire

func newBloodswornSquire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bloodsworn Squire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{B}, Discard a card: This creature gains indestructible until end of turn. Tap it. Then if there are four or more creature cards in your graveyard, transform this creature. (Damage and effects that say \"destroy\" don't destroy it.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalDiscard,
							Text:        "Discard a card",
							Amount:      1,
							Source:      zone.Hand,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Indestructible},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.Tap{
											Object: game.SourcePermanentReference(),
										},
									},
									game.Instruction{
										Primitive: game.Transform{
											Object: game.SourcePermanentReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "if there are four or more creature cards in your graveyard",
												SourceCounterKind: counter.PlusOnePlusOne,
												ControllerGraveyardCardOfTypeCountAtLeast: 4,
												ControllerGraveyardCountCardType:          types.Creature,
											}),
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
			{1}{B}, Discard a card: This creature gains indestructible until end of turn. Tap it. Then if there are four or more creature cards in your graveyard, transform this creature. (Damage and effects that say "destroy" don't destroy it.)
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:             "Bloodsworn Knight",
			Colors:           []color.Color{color.Black},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Vampire, types.Knight},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCardsInGraveyard}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCardsInGraveyard}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{B}, Discard a card: This creature gains indestructible until end of turn. Tap it. (Damage and effects that say \"destroy\" don't destroy it.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalDiscard,
							Text:        "Discard a card",
							Amount:      1,
							Source:      zone.Hand,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.SourcePermanentReference()),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Indestructible},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.Tap{
											Object: game.SourcePermanentReference(),
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
			Bloodsworn Knight's power and toughness are each equal to the number of creature cards in your graveyard.
			{1}{B}, Discard a card: This creature gains indestructible until end of turn. Tap it. (Damage and effects that say "destroy" don't destroy it.)
		`,
		}),
	}
}
