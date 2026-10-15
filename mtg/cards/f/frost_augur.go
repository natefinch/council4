package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FrostAugur is the card definition for Frost Augur.
//
// Type: Snow Creature — Human Wizard
// Cost: {U}
//
// Oracle text:
//
//	{S}, {T}: Look at the top card of your library. If it's a snow card, you may reveal it and put it into your hand. ({S} can be paid with one mana from a snow source.)
var FrostAugur = newFrostAugur

func newFrostAugur() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Frost Augur",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Wizard},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{S}, {T}: Look at the top card of your library. If it's a snow card, you may reveal it and put it into your hand. ({S} can be paid with one mana from a snow source.)",
					ManaCost: opt.Val(cost.Mana{
						cost.S,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LookAtLibraryTop{
											Player:        game.ControllerReference(),
											PublishLinked: game.LinkedKey("sequence-effect-0-product"),
										},
										LocalProducts: game.LocalProducts{
											Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
										},
									},
									game.Instruction{
										Primitive: game.Reveal{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "sequence-effect-0-product",
											},
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:              "If it's a snow card",
												Object:            opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
												ObjectMatches:     opt.Val(game.Selection{Supertypes: []types.Super{types.Snow}}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
										PublishCondition:        game.ConditionKey("condition-0"),
										Optional:                true,
										PublishOptionalDecision: game.OptionalDecisionKey("optional-clause-2"),
									},
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "sequence-effect-0-product",
											},
											FromZone:    zone.Library,
											Destination: zone.Hand,
										},
										ConditionGate:        game.ConditionKey("condition-0"),
										OptionalDecisionGate: game.OptionalDecisionKey("optional-clause-2"),
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
			{S}, {T}: Look at the top card of your library. If it's a snow card, you may reveal it and put it into your hand. ({S} can be paid with one mana from a snow source.)
		`,
		},
	}
}
