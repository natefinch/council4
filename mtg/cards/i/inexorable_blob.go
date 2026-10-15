package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InexorableBlob is the card definition for Inexorable Blob.
//
// Type: Creature — Ooze
// Cost: {2}{G}
//
// Oracle text:
//
//	Delirium — Whenever this creature attacks, if there are four or more card types among cards in your graveyard, create a 3/3 green Ooze creature token that's tapped and attacking.
var InexorableBlob = newInexorableBlob

func newInexorableBlob() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Inexorable Blob",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ooze},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Delirium — Whenever this creature attacks, if there are four or more card types among cards in your graveyard, create a 3/3 green Ooze creature token that's tapped and attacking.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if there are four or more card types among cards in your graveyard",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if there are four or more card types among cards in your graveyard",
							Aggregates: []game.AggregateComparison{
								game.AggregateComparison{
									Aggregate: game.AggregateControllerGraveyardCardTypeCount,
									Op:        compare.GreaterOrEqual,
									Value:     4,
								},
							},
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:         game.Fixed(1),
											Source:         game.TokenDef(inexorableBlobToken),
											EntryTapped:    true,
											EntryAttacking: true,
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
			Delirium — Whenever this creature attacks, if there are four or more card types among cards in your graveyard, create a 3/3 green Ooze creature token that's tapped and attacking.
		`,
		},
	}
}

var inexorableBlobToken = newInexorableBlobToken()

func newInexorableBlobToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Ooze",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ooze},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
