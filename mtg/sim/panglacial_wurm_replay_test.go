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

type searchCastReplayPolicy struct {
	castTutor, castWurm bool
}

func (p *searchCastReplayPolicy) ChooseAction(obs rules.PlayerObservation, legal []action.Action) action.Action {
	for _, candidate := range legal {
		if cast, ok := candidate.CastSpellPayload(); ok {
			for _, card := range obs.Hand() {
				if !p.castTutor && card.CardInstanceID == cast.CardID && card.Name == "Demonic Tutor" {
					p.castTutor = true
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

func (p *searchCastReplayPolicy) ChooseChoice(_ rules.PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Cast a card before finding cards?" {
		if !p.castWurm {
			p.castWurm = true
			return []int{1}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

func TestPanglacialWurmSearchCastsReplayThroughJSON(t *testing.T) {
	// Rules scenarios cover printed payment; this fixture isolates search replay.
	wurm, tutor := authored.PanglacialWurm.New(), authored.DemonicTutor.New()
	wurm.ManaCost, tutor.ManaCost = opt.Val(cost.Mana{}), opt.Val(cost.Mana{})
	cfg := smokeConfig(1, 3286)
	cfg.TurnLimit = 1
	cfg.Configs[game.Player1].Deck = nil
	for range 8 {
		cfg.Configs[game.Player1].Deck = append(cfg.Configs[game.Player1].Deck, wurm, tutor)
	}
	cfg.NewAgents = func(uint64) [game.NumPlayers]rules.PlayerAgent {
		var policies [game.NumPlayers]rules.PlayerAgent
		for player := range policies {
			policies[player] = &searchCastReplayPolicy{}
		}
		return policies
	}
	result, record := RecordGame(cfg, 0)
	wurms, tutors := 0, 0
	for _, turn := range result.Turns {
		for _, entry := range turn.Entries {
			if entry.Kind == rules.TurnLogEntryResolve {
				switch entry.Resolve.SourceName {
				case "Panglacial Wurm":
					wurms++
				case "Demonic Tutor":
					tutors++
				}
			}
		}
	}
	if wurms != 1 || tutors != 1 || result.Execution.InvalidActions != 0 || result.Execution.InvalidChoices != 0 {
		t.Fatalf("fixture must resolve one tutor and one search-cast Wurm legally, got tutors=%d wurms=%d diagnostics=%+v",
			tutors, wurms, result.Execution)
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
