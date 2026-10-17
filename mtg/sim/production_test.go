package sim

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

type invalidProductionAgent struct{}

func (invalidProductionAgent) ChooseAction(rules.PlayerObservation, []action.Action) action.Action {
	return action.CastSpell(999999, nil, 0, nil)
}

func productionConfig(t *testing.T) Config {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	cfg := Config{Games: 2, Seed: 20261014, TurnLimit: 12, Workers: 1}
	for p := range cfg.Configs {
		for range 70 {
			cfg.Configs[p].Deck = append(cfg.Configs[p].Deck, registry.Lookup("Forest"))
		}
		for _, entry := range authored.Cards {
			def, err := registry.LookupOracleID(entry.OracleID)
			if err != nil {
				t.Fatal(err)
			}
			cfg.Configs[p].Deck = append(cfg.Configs[p].Deck, def)
		}
	}
	return cfg
}

func TestMixedProductionBatchReplayAndSerializedPrograms(t *testing.T) {
	cfg := productionConfig(t)
	serial := Run(cfg)
	cfg.Workers = 2
	parallel := Run(cfg)
	if !reflect.DeepEqual(serial, parallel) || serial.FailureCount() != 0 {
		t.Fatal("mixed production workers diverged or failed")
	}
	result, record := RecordGame(cfg, 0)
	if !reflect.DeepEqual(result, serial.Games[0]) || record.TurnLimit != cfg.TurnLimit || len(record.Programs) == 0 {
		t.Fatal("recording changed configured work or lost program versions")
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var reload ReplayRecord
	if err := json.Unmarshal(data, &reload); err != nil {
		t.Fatal(err)
	}
	if got := Replay(cfg.Configs, reload); !reflect.DeepEqual(got, result) {
		t.Fatal("mixed production serialization/reload changed executed result")
	}
	reload.Programs[0].Version = "different-version"
	defer func() {
		if _, ok := recover().(rules.UnsupportedError); !ok {
			t.Fatal("changed program version was accepted for replay")
		}
	}()
	Replay(cfg.Configs, reload)
}

func TestProductionOutcomeCapsAndInvalidFallbackAreVisible(t *testing.T) {
	cfg := productionConfig(t)
	cfg.Games, cfg.TurnLimit = 1, 1
	cfg.NewAgents = func(uint64) [game.NumPlayers]rules.PlayerAgent {
		var agents [game.NumPlayers]rules.PlayerAgent
		for i := range agents {
			agents[i] = invalidProductionAgent{}
		}

		return agents
	}

	result := Run(cfg)
	if result.CompletedCount() != 0 || result.DrawCount() != 0 || !result.Games[0].TurnLimitReached ||
		result.Games[0].Execution.InvalidActions == 0 || result.Games[0].Execution.ActionFallbacks == 0 ||
		result.Games[0].Exact() || len(result.Games[0].Limitations) == 0 {
		t.Fatalf("cap/fallback/coverage misreported: %+v", result.Games[0])
	}
	cfg.DecisionLimit = 1
	result = Run(cfg)
	if result.Games[0].Outcome() != rules.TerminationWorkBudget || result.Games[0].Execution.Decisions != 1 ||
		result.DrawCount() != 0 || result.CompletedCount() != 0 || result.FailureCount() != 1 || result.Failures[0].Unsupported {
		t.Fatal("work exhaustion was counted as a draw or unsupported mechanic")
	}
}

func TestInvalidActionReplayRetainsFallbackDiagnostics(t *testing.T) {
	cfg := productionConfig(t)
	cfg.TurnLimit = 1
	cfg.NewAgents = func(uint64) [game.NumPlayers]rules.PlayerAgent {
		var agents [game.NumPlayers]rules.PlayerAgent
		for i := range agents {
			agents[i] = invalidProductionAgent{}
		}
		return agents
	}
	result, record := RecordGame(cfg, 0)
	if result.Execution.InvalidActions == 0 {
		t.Fatal("fixture did not generate an invalid answer")
	}
	if replayed := Replay(cfg.Configs, record); !reflect.DeepEqual(replayed, result) {
		t.Fatalf("fallback status changed during replay: %+v != %+v", replayed.Execution, result.Execution)
	}
}

func TestStaticOnlyReplayNeedsNoPrograms(t *testing.T) {
	cfg := productionConfig(t)
	registry := cards.NewDefaultRegistry()
	for p := range cfg.Configs {
		cfg.Configs[p].Deck = []*game.CardDef{registry.Lookup("Blood Moon")}
		for range 30 {
			cfg.Configs[p].Deck = append(cfg.Configs[p].Deck, registry.Lookup("Forest"))
		}
	}
	cfg.TurnLimit = 1
	result, record := RecordGame(cfg, 0)
	if len(record.Programs) != 0 {
		t.Fatal("static-only deck recorded a resolving program")
	}
	if replayed := Replay(cfg.Configs, record); !reflect.DeepEqual(replayed, result) {
		t.Fatal("static-only replay changed the result")
	}
}
