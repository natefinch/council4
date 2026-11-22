package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	bloodbornScoundrelsOracleID   = "37c1de8d-30d6-4b08-9368-279cf66d3e85"
	bloodbornScoundrelsOracleText = "Assist (Another player can pay up to {5} of this spell's cost.)\n" +
		"When this creature enters, target opponent loses 2 life and you gain 2 life."
	bloodbornScoundrelsEnterText = "When this creature enters, target opponent loses 2 life and you gain 2 life."
)

var BloodbornScoundrels = cardset.Entry{
	OracleID: bloodbornScoundrelsOracleID,
	Name:     "Bloodborn Scoundrels",
	New:      newBloodbornScoundrels,
	Verification: "mtg/rules/bloodborn_scoundrels_test.go: " +
		"TestBloodbornScoundrelsAssistAndEnterLifeTrigger",
}

func newBloodbornScoundrels() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Bloodborn Scoundrels",
			ManaCost:   opt.Val(cost.Mana{cost.O(5), cost.B}),
			OracleText: bloodbornScoundrelsOracleText,
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Vampire, types.Rogue},
			Colors:     []color.Color{color.Black},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.AssistStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: bloodbornScoundrelsEnterText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{
					Text: bloodbornScoundrelsEnterText,
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					}},
					Sequence: []game.Instruction{
						{Primitive: game.LoseLife{Player: game.TargetPlayerReference(0), Amount: game.Fixed(2)}},
						{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)}},
					},
				}.Ability(),
			}},
		},
	}
}
