package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PsychotropeThallid is the card definition for Psychotrope Thallid.
//
// Type: Creature — Fungus
// Cost: {2}{G}
//
// Oracle text:
//
//	At the beginning of your upkeep, put a spore counter on this creature.
//	Remove three spore counters from this creature: Create a 1/1 green Saproling creature token.
//	{1}, Sacrifice a Saproling: Draw a card.
var PsychotropeThallid = newPsychotropeThallid

func newPsychotropeThallid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Psychotrope Thallid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fungus},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove three spore counters from this creature: Create a 1/1 green Saproling creature token.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove three spore counters from this creature",
							Amount:      3,
							CounterKind: counter.Spore,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(psychotropeThallidToken),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:     "{1}, Sacrifice a Saproling: Draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a Saproling",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Saproling},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
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
			Remove three spore counters from this creature: Create a 1/1 green Saproling creature token.
			{1}, Sacrifice a Saproling: Draw a card.
		`,
		},
	}
}

var psychotropeThallidToken = newPsychotropeThallidToken()

func newPsychotropeThallidToken() *game.CardDef {
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
