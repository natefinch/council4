package sim

import (
	"encoding/json"
	"reflect"
	"slices"
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

type replicateReplayPolicy struct{}

func (replicateReplayPolicy) ChooseAction(obs rules.PlayerObservation, candidates []action.Action) action.Action {
	for _, candidate := range candidates {
		if cast, ok := candidate.CastSpellPayload(); ok {
			// Each instance is sampled independently. Force an exact proposal
			// rather than a candidate index so JSON must preserve both payments.
			cast.ReplicateCounts = []int{2, 1}
			return action.AnnounceSpell(cast)
		}
	}
	return (agent.Agent{Strategy: agent.GenericStrategy{}}).ChooseAction(obs, candidates)
}

func TestReplicateAnnouncementsReplayThroughJSON(t *testing.T) {
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Replicate replay fixture", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.O(0)}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{
				game.ReplicateKeyword{Cost: cost.Mana{cost.O(0)}},
				game.ReplicateKeyword{Cost: cost.Mana{cost.O(0)}},
			},
		}},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
	cfg := Config{Seed: 4028, Games: 1, TurnLimit: 1, NewAgents: func(uint64) [game.NumPlayers]rules.PlayerAgent {
		var policies [game.NumPlayers]rules.PlayerAgent
		for i := range policies {
			policies[i] = replicateReplayPolicy{}
		}
		return policies
	}}
	for i := range cfg.Configs {
		for range 8 {
			cfg.Configs[i].Deck = append(cfg.Configs[i].Deck, def)
		}
	}
	result, record := RecordGame(cfg, 0)
	found := false
	for _, seat := range record.Seats {
		for _, proposal := range seat.Announcements {
			found = found || proposal.Cast != nil && slices.Equal(proposal.Cast.ReplicateCounts, []int{2, 1})
		}
	}
	if !found || result.Execution.InvalidActions != 0 || result.Execution.InvalidChoices != 0 {
		t.Fatal("recording omitted or rejected the exact Replicate announcement")
	}
	casts, copies := 0, 0
	for _, event := range result.Events {
		switch event.Kind {
		case game.EventSpellCast:
			casts++
		case game.EventSpellCopied:
			copies++
		}
	}
	if casts == 0 || copies != casts*3 {
		t.Fatalf("recorded %d casts and %d copies, want three copies per cast", casts, copies)
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	var restored ReplayRecord
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if replayed := Replay(cfg.Configs, restored); !reflect.DeepEqual(result, replayed) {
		t.Fatal("Replicate announcements diverged in actual JSON replay")
	}
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
