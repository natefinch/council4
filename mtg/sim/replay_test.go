package sim

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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

// TestReplayWithDefaultAgentsAndTutorChoice guards the MinChoices==0 case: the
// default FirstLegal agents are not ChoiceAgents, so the engine answers the tutor
// (search) choice with its fallback. A non-choice-capable seat must replay
// through the same fallback rather than answering nil (a valid empty selection
// for a search), which would otherwise find nothing and diverge.
func TestReplayWithDefaultAgentsAndTutorChoice(t *testing.T) {
	cfg := Config{Configs: tutorConfigs(), Games: 1, Seed: 314} // nil NewAgents -> FirstLegal
	result, record := RecordGame(cfg, 0)

	for _, seat := range record.Seats {
		if seat.ChoiceCapable {
			t.Fatal("a FirstLegal seat should not be marked choice-capable")
		}
	}
	if replayed := Replay(cfg.Configs, record); !reflect.DeepEqual(replayed, result) {
		t.Error("replay with default agents and a tutor (MinChoices==0) choice diverged from the recording")
	}
}

// tutorConfigs builds decks of free "search your library for a creature"
// sorceries plus creatures to find, so a non-ChoiceAgent seat exercises the
// engine's fallback handling of a MinChoices==0 search choice.
func tutorConfigs() [game.NumPlayers]game.PlayerConfig {
	tutor := &game.CardDef{CardFace: game.CardFace{
		Name:  "Free Tutor",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Search{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
				Spec: game.SearchSpec{
					SourceZone:  zone.Library,
					Destination: zone.Hand,
					Filter: game.Selection{
						RequiredTypes: []types.Card{types.Creature},
					},
				},
			},
		}}}.Ability()),
	}}
	bear := &game.CardDef{CardFace: game.CardFace{
		Name:      "Findable Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
	var configs [game.NumPlayers]game.PlayerConfig
	for player := range configs {
		for range 8 {
			configs[player].Deck = append(configs[player].Deck, tutor)
		}
		for range 4 {
			configs[player].Deck = append(configs[player].Deck, bear)
		}
	}
	return configs
}
