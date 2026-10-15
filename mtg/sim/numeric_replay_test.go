package sim

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/agent"
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

type numericReplayPolicy struct{}

func (numericReplayPolicy) ChooseAction(obs rules.PlayerObservation, candidates []action.Action) action.Action {
	for _, candidate := range candidates {
		if cast, ok := candidate.CastSpellPayload(); ok {
			cast.KickerPaid, cast.KickerCount = true, 21
			return action.AnnounceSpell(cast)
		}
	}
	return (agent.Agent{Strategy: agent.GenericStrategy{}}).ChooseAction(obs, candidates)
}

func (numericReplayPolicy) ChooseChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceNumber {
		return []int{21}
	}
	return (agent.BaselineStrategy{}).ChooseChoice(obs, request)
}

func TestExactNumericAnnouncementsAndChoicesReplayThroughJSON(t *testing.T) {
	reference := game.ProgramReference{ID: "test/numeric-replay", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference, Resolve: func(ctx cardprogram.Context) error {
			chosen, err := ctx.Ask(game.ChoiceRequest{
				Kind: game.ChoiceNumber, Player: ctx.View().Source().Controller,
				NumberRange: opt.Val(game.NumberRange{Minimum: 0}),
				MinChoices:  1, MaxChoices: 1, DefaultSelection: []int{0},
			})
			if err != nil {
				return err
			}
			_, err = ctx.Execute(game.GainLife{Amount: game.Fixed(chosen[0]), Player: game.ControllerReference()})
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Exact numeric replay", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(0)}, Multi: true}},
		}},
		SpellAbility: opt.Val(game.Mode{Program: reference}.Ability()),
	}}
	cfg := Config{Seed: 3284, Games: 1, Programs: &catalog, TurnLimit: 2,
		NewAgents: func(uint64) [game.NumPlayers]rules.PlayerAgent {
			var policies [game.NumPlayers]rules.PlayerAgent
			for index := range policies {
				policies[index] = numericReplayPolicy{}
			}
			return policies
		},
	}
	for player := range cfg.Configs {
		for range 12 {
			cfg.Configs[player].Deck = append(cfg.Configs[player].Deck, def)
		}
	}
	result, record := RecordGame(cfg, 0)
	announcements, numbers := 0, 0
	for _, seat := range record.Seats {
		for _, proposal := range seat.Announcements {
			if proposal.Cast != nil && proposal.Cast.KickerCount == 21 {
				announcements++
			}
		}
		for _, choice := range seat.Choices {
			if len(choice) == 1 && choice[0] == 21 {
				numbers++
			}
		}
	}
	if announcements == 0 || numbers == 0 || result.Execution.InvalidActions != 0 || result.Execution.InvalidChoices != 0 {
		t.Fatal("recording omitted exact actions/numbers or treated them as invalid")
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var restored ReplayRecord
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if replayed := Replay(cfg.Configs, restored, catalog); !reflect.DeepEqual(result, replayed) {
		t.Fatal("exact announcements or scalar choices diverged on actual replay")
	}
}
