package sim

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

type searchCastReplayPolicy struct{}

func (searchCastReplayPolicy) ChooseAction(obs rules.PlayerObservation, legal []action.Action) action.Action {
	for _, candidate := range legal {
		if cast, ok := candidate.CastSpellPayload(); ok {
			for _, card := range obs.Hand() {
				if card.CardInstanceID == cast.CardID && card.Name == "Demonic Tutor" {
					return candidate
				}
			}
		}
		if _, ok := candidate.DeclareAttackersPayload(); ok {
			return candidate
		}
		if _, ok := candidate.DeclareBlockersPayload(); ok {
			return candidate
		}
	}
	return action.Pass()
}

func (searchCastReplayPolicy) ChooseChoice(_ rules.PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Cast a card before finding cards?" {
		return []int{1}
	}
	return request.DefaultSelection
}

func TestPanglacialWurmSearchCastsReplayThroughJSON(t *testing.T) {
	// Rules scenarios cover printed payment; this fixture isolates search replay.
	wurm, tutor := authored.PanglacialWurm.New(), authored.DemonicTutor.New()
	wurm.ManaCost, tutor.ManaCost = opt.Val(cost.Mana{}), opt.Val(cost.Mana{})
	cfg := Config{Seed: 3286, Games: 1, TurnLimit: 4}
	for player := range cfg.Configs {
		for range 15 {
			cfg.Configs[player].Deck = append(cfg.Configs[player].Deck, wurm, tutor)
		}
	}
	cfg.NewAgents = func(uint64) [game.NumPlayers]rules.PlayerAgent {
		var policies [game.NumPlayers]rules.PlayerAgent
		for player := range policies {
			policies[player] = searchCastReplayPolicy{}
		}
		return policies
	}
	result, record := RecordGame(cfg, 0)
	resolved := false
	for _, turn := range result.Turns {
		for _, entry := range turn.Entries {
			resolved = resolved || entry.Resolve.SourceName == "Panglacial Wurm"
		}
	}
	if !resolved || result.Execution.InvalidActions != 0 || result.Execution.InvalidChoices != 0 {
		t.Fatal("record did not execute a legal Wurm cast (policy never casts it from hand)")
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var restored ReplayRecord
	if err := json.Unmarshal(encoded, &restored); err != nil {
		t.Fatal(err)
	}
	if replayed := Replay(cfg.Configs, restored); !reflect.DeepEqual(replayed, result) {
		t.Fatal("search-cast choices and stack order diverged after JSON replay")
	}
	if repeated := RunOne(cfg, 0); !reflect.DeepEqual(repeated, result) {
		t.Fatal("same-input search-cast execution was not reproducible")
	}
}
