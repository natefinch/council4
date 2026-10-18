package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const marchesaDealerOfDeathOracleID = "37449c93-4d32-4f1a-982f-8f3416214368"

var MarchesaDealerOfDeath = cardset.Entry{
	OracleID: marchesaDealerOfDeathOracleID,
	Name:     "Marchesa, Dealer of Death",
	Verification: "mtg/rules/marchesa_dealer_of_death_test.go: " +
		"TestMarchesaDealerOfDeathCrimeTriggerPaysAndResolvesAfterSourceLeaves, " +
		"TestMarchesaDealerOfDeathDeclinesOrCannotPay, " +
		"TestMarchesaDealerOfDeathOneCardLibrary, " +
		"TestMarchesaDealerOfDeathDoesNotTriggerForCrimeAgainstSelf, " +
		"TestMarchesaDealerOfDeathTriggersForActivatedAndTriggeredCrimes",
	New: newMarchesaDealerOfDeath,
}

func newMarchesaDealerOfDeath() *game.CardDef {
	text := "Whenever you commit a crime, you may pay {1}. If you do, look at the top two cards of your library. Put one of them into your hand and the other into your graveyard. (Targeting opponents, anything they control, and/or cards in their graveyards is a crime.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name:       "Marchesa, Dealer of Death",
			ManaCost:   opt.Val(cost.Mana{cost.U, cost.B, cost.R}),
			Colors:     []color.Color{color.Blue, color.Black, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Rogue},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			OracleText: text,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: text,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:  game.EventCrimeCommitted,
						Player: game.TriggerPlayerYou,
					},
				},
				Content: game.AbilityContent{
					Modes: []game.Mode{{Sequence: []game.Instruction{
						{
							Primitive: game.Pay{Payment: game.ResolutionPayment{
								Prompt:   "Pay {1}?",
								ManaCost: opt.Val(cost.Mana{cost.O(1)}),
							}},
							PublishResult: "paid",
						},
						{
							Primitive: game.Dig{
								Player:    game.ControllerReference(),
								Look:      game.Fixed(2),
								Take:      game.Fixed(1),
								Remainder: game.DigRemainderGraveyard,
							},
							ResultGate: opt.Val(game.InstructionResultGate{
								Key:       "paid",
								Succeeded: game.TriTrue,
							}),
						},
					}}},
					MinModes: 1,
					MaxModes: 1,
				},
			}},
		},
	}
}
