package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SporolothAncient is the card definition for Sporoloth Ancient.
//
// Type: Creature — Fungus
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	At the beginning of your upkeep, put a spore counter on this creature.
//	Creatures you control have "Remove two spore counters from this creature: Create a 1/1 green Saproling creature token."
var SporolothAncient = newSporolothAncient

func newSporolothAncient() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sporoloth Ancient",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fungus},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							AddAbilities: []game.Ability{
								new(game.ActivatedAbility{
									Text: "Remove two spore counters from this creature: Create a 1/1 green Saproling creature token.",
									AdditionalCosts: []cost.Additional{
										{
											Kind:        cost.AdditionalRemoveCounter,
											Text:        "Remove two spore counters from this creature",
											Amount:      2,
											CounterKind: counter.Spore,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.CreateToken{
													Amount: game.Fixed(1),
													Source: game.TokenDef(sporolothAncientToken),
												},
											},
										},
									}.Ability(),
								}),
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
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Spore,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your upkeep, put a spore counter on this creature.
			Creatures you control have "Remove two spore counters from this creature: Create a 1/1 green Saproling creature token."
		`,
		},
	}
}

var sporolothAncientToken = newSporolothAncientToken()

func newSporolothAncientToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Saproling",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Saproling},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
