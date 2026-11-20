package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const trainOfThoughtText = "Replicate {1}{U} (When you cast this spell, copy it for each time you paid its replicate cost.)\nDraw a card."

var TrainOfThoughtProgram = game.ProgramReference{
	ID:      "90ba0366-57ff-4dc4-a290-3925548129ce/spell/0",
	Version: "1",
}

var TrainOfThought = cardset.Entry{
	OracleID: "90ba0366-57ff-4dc4-a290-3925548129ce",
	Name:     "Train of Thought",
	New:      newTrainOfThought,
	Programs: []cardprogram.Program{{
		Reference: TrainOfThoughtProgram,
		Resolve:   resolveTrainOfThought,
	}},
	Verification: "mtg/rules/train_of_thought_test.go: " +
		"TestAuthoredTrainOfThoughtReplicateDrawsForEachPayment",
}

func newTrainOfThought() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Train of Thought",
			OracleText: trainOfThoughtText,
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.U}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Blue},
			StaticAbilities: []game.StaticAbility{{
				KeywordAbilities: []game.KeywordAbility{
					game.ReplicateKeyword{Cost: cost.Mana{cost.O(1), cost.U}},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text:    "Draw a card.",
				Program: TrainOfThoughtProgram,
			}.Ability()),
		},
	}
}

func resolveTrainOfThought(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.Draw{
		Player: game.ControllerReference(),
		Amount: game.Fixed(1),
	})
	return err
}
