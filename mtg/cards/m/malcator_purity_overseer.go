package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MalcatorPurityOverseer is the card definition for Malcator, Purity Overseer.
//
// Type: Legendary Creature — Phyrexian Elephant Wizard
// Cost: {1}{W}{U}
//
// Oracle text:
//
//	When Malcator enters, create a 3/3 colorless Phyrexian Golem artifact creature token.
//	At the beginning of your end step, if three or more artifacts entered the battlefield under your control this turn, create a 3/3 colorless Phyrexian Golem artifact creature token.
var MalcatorPurityOverseer = newMalcatorPurityOverseer()

func newMalcatorPurityOverseer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Malcator, Purity Overseer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Elephant, types.Wizard},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
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
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(malcatorPurityOverseerToken),
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
						InterveningIf: "if three or more artifacts entered the battlefield under your control this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:            game.EventPermanentEnteredBattlefield,
								Controller:       game.TriggerControllerYou,
								SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							}, Window: game.EventHistoryCurrentTurn, MinCount: 3}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(malcatorPurityOverseerToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When Malcator enters, create a 3/3 colorless Phyrexian Golem artifact creature token.
			At the beginning of your end step, if three or more artifacts entered the battlefield under your control this turn, create a 3/3 colorless Phyrexian Golem artifact creature token.
		`,
		},
	}
}

var malcatorPurityOverseerToken = newMalcatorPurityOverseerToken()

func newMalcatorPurityOverseerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Phyrexian Golem",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
