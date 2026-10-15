package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GenesisChamber is the card definition for Genesis Chamber.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Whenever a nontoken creature enters, if this artifact is untapped, that creature's controller creates a 1/1 colorless Myr artifact creature token.
var GenesisChamber = newGenesisChamber

func newGenesisChamber() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Genesis Chamber",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a nontoken creature enters, if this artifact is untapped, that creature's controller creates a 1/1 colorless Myr artifact creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, NonToken: true},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
						InterveningIf: "if this artifact is untapped",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if this artifact is untapped",
							Object:            opt.Val(game.SourcePermanentReference()),
							ObjectMatches:     opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Tapped: game.TriFalse}),
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:    game.Fixed(1),
											Source:    game.TokenDef(genesisChamberToken),
											Recipient: opt.Val(game.ObjectControllerReference(game.EventPermanentReference())),
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
			Whenever a nontoken creature enters, if this artifact is untapped, that creature's controller creates a 1/1 colorless Myr artifact creature token.
		`,
		},
	}
}

var genesisChamberToken = newGenesisChamberToken()

func newGenesisChamberToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Myr",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Myr},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
