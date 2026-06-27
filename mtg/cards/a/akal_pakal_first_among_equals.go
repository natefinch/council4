package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AkalPakalFirstAmongEquals is the card definition for Akal Pakal, First Among Equals.
//
// Type: Legendary Creature — Human Advisor
// Cost: {2}{U}
//
// Oracle text:
//
//	At the beginning of each player's end step, if an artifact entered the battlefield under your control this turn, look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.
var AkalPakalFirstAmongEquals = newAkalPakalFirstAmongEquals()

func newAkalPakalFirstAmongEquals() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Akal Pakal, First Among Equals",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Advisor},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepEnd,
						},
						InterveningIf: "if an artifact entered the battlefield under your control this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:            game.EventPermanentEnteredBattlefield,
								Controller:       game.TriggerControllerYou,
								SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player: game.ControllerReference(),
									Look:   game.Fixed(2),
									Take:   game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of each player's end step, if an artifact entered the battlefield under your control this turn, look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.
		`,
		},
	}
}
