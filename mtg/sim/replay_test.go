package sim

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

// genericAgents seats a GenericStrategy agent at every seat. The agent answers
// engine-mediated choices (it is a rules.ChoiceAgent), so recording exercises the
// choice path — land-only smoke decks overfill their hands and face cleanup
// discard choices.
func genericAgents(uint64) [game.NumPlayers]rules.PlayerAgent {
	var agents [game.NumPlayers]rules.PlayerAgent
	for i := range agents {
		agents[i] = agent.Agent{Strategy: agent.GenericStrategy{}}
	}
	return agents
}

func recordableConfig(seed uint64) Config {
	cfg := smokeConfig(1, seed)
	cfg.NewAgents = genericAgents
	return cfg
}

func TestRecordedGameReplaysIdentically(t *testing.T) {
	cfg := recordableConfig(2024)
	result, record := RecordGame(cfg, 0)

	// The recorded result matches a plain run of the same game.
	if !reflect.DeepEqual(result, RunOne(cfg, 0)) {
		t.Fatal("RecordGame result differs from RunOne for the same game")
	}
	// Replaying the record reproduces the same result.
	replayed := Replay(cfg.Configs, record)
	if !reflect.DeepEqual(replayed, result) {
		t.Error("Replay did not reproduce the recorded GameResult")
	}
}

func TestReplayRecordRoundTripsThroughJSON(t *testing.T) {
	cfg := recordableConfig(77)
	result, record := RecordGame(cfg, 0)

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal record: %v", err)
	}
	var decoded ReplayRecord
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal record: %v", err)
	}
	if decoded.Seed != record.Seed {
		t.Errorf("round-tripped seed = %d, want %d", decoded.Seed, record.Seed)
	}
	if replayed := Replay(cfg.Configs, decoded); !reflect.DeepEqual(replayed, result) {
		t.Error("replay from a JSON round-tripped record did not reproduce the result")
	}
}

func TestRecordCapturesChoices(t *testing.T) {
	// Free scry sorceries resolve a scry each cast, an engine-mediated choice,
	// so every seat makes scry choices that
	// recording must capture.
	cfg := Config{Configs: scryDeckConfigs(), Games: 1, Seed: 9, NewAgents: genericAgents}
	_, record := RecordGame(cfg, 0)
	captured := false
	for _, seat := range record.Seats {
		if len(seat.Choices) > 0 {
			captured = true
		}
		if len(seat.Actions) == 0 {
			t.Error("a seat recorded no actions; expected at least priority passes")
		}
	}
	if !captured {
		t.Error("expected at least one recorded choice (cleanup discard) across seats")
	}
}

// scryDeckConfigs builds four decks of free scry sorceries. Each cast resolves a
// scry, an engine-mediated choice the agents answer and recording must capture.
func scryDeckConfigs() [game.NumPlayers]game.PlayerConfig {
	scry := &game.CardDef{CardFace: game.CardFace{
		Name:  "Free Scry",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{
			{Primitive: game.Scry{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		}}.Ability()),
	}}
	var configs [game.NumPlayers]game.PlayerConfig
	for player := range configs {
		for range 12 {
			configs[player].Deck = append(configs[player].Deck, scry)
		}
	}
	return configs
}

func TestReplayIsDeterministic(t *testing.T) {
	cfg := recordableConfig(555)
	_, record := RecordGame(cfg, 0)
	first := Replay(cfg.Configs, record)
	second := Replay(cfg.Configs, record)
	if !reflect.DeepEqual(first, second) {
		t.Error("Replay is not deterministic for the same record")
	}
}
