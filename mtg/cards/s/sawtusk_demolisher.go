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

// SawtuskDemolisher is the card definition for Sawtusk Demolisher.
//
// Type: Creature — Beast
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Mutate {3}{G} (If you cast this spell for its mutate cost, put it over or under target non-Human creature you own. They mutate into the creature on top plus all abilities from under it.)
//	Trample
//	Whenever this creature mutates, destroy target noncreature permanent. Its controller creates a 3/3 green Beast creature token.
var SawtuskDemolisher = newSawtuskDemolisher

func newSawtuskDemolisher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sawtusk Demolisher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text:           "Mutate {3}{G}",
					ZoneOfFunction: zone.Hand,
					KeywordAbilities: []game.KeywordAbility{
						game.MutateKeyword{
							Cost: cost.Mana{
								cost.O(3),
								cost.G,
							},
						},
					},
				},
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature mutates, destroy target noncreature permanent. Its controller creates a 3/3 green Beast creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentMutated,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target noncreature permanent",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:    game.Fixed(1),
											Source:    game.TokenDef(sawtuskDemolisherToken),
											Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),
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
			Mutate {3}{G} (If you cast this spell for its mutate cost, put it over or under target non-Human creature you own. They mutate into the creature on top plus all abilities from under it.)
			Trample
			Whenever this creature mutates, destroy target noncreature permanent. Its controller creates a 3/3 green Beast creature token.
		`,
		},
	}
}

var sawtuskDemolisherToken = newSawtuskDemolisherToken()

func newSawtuskDemolisherToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
