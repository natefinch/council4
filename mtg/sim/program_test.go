package sim

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

func authoredSimulation(t *testing.T) Config {
	t.Helper()
	ref := game.ProgramReference{ID: "test/gain-life", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: ref, Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Execute(game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()})
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	def := &game.CardDef{CardFace: game.CardFace{Name: "Authored Gain", Types: []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{}), SpellAbility: opt.Val(game.Mode{Program: ref}.Ability())}}
	cfg := Config{Seed: 7, Games: 2, Programs: &catalog, TurnLimit: 1}
	for player := range game.NumPlayers {
		for range 8 {
			cfg.Configs[player].Deck = append(cfg.Configs[player].Deck, def)
		}
	}
	return cfg
}

func TestAuthoredParallelGamesAndReplayShareOnlyImmutableCode(t *testing.T) {
	t.Parallel()
	cfg := authoredSimulation(t)
	cfg.Workers = 1
	serial := Run(cfg)
	cfg.Workers = 2
	parallel := Run(cfg)
	if len(serial.Failures) != 0 || !reflect.DeepEqual(serial, parallel) {
		t.Fatal("parallel authored games diverged or failed")
	}
	original, record := RecordGame(cfg, 0)
	if original.Outcome() != rules.TerminationTurnBudget || original.EndState.Players[game.Player1].Life <= 40 {
		t.Fatal("one-turn fixture did not execute the authored gain-life body")
	}
	if len(record.Programs) != 1 {
		t.Fatal("replay omitted the code version")
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var restored ReplayRecord
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	replayed := Replay(cfg.Configs, restored, *cfg.Programs)
	if !reflect.DeepEqual(original, replayed) {
		t.Fatal("authored replay did not reproduce its structured result")
	}
}

func TestMissingProgramFailsBeforeARealGameIsReported(t *testing.T) {
	t.Parallel()
	cfg := authoredSimulation(t)
	cfg.Programs = nil
	result := Run(cfg)
	if len(result.Failures) != cfg.Games || !result.Failures[0].Unsupported {
		t.Fatal("missing code was admitted or hidden")
	}
	for _, played := range result.Games {
		if played.TurnCount != 0 || len(played.Events) != 0 || played.HasWinner ||
			played.Outcome() != rules.TerminationUnsupported || played.Completed() {
			t.Fatal("preflight failure was reported as an executed game")
		}
	}
}
