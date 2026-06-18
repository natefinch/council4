package agent

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

// scriptedStrategy scores a chosen action kind highest and records the choices
// it is asked to make.
type scriptedStrategy struct {
	preferKind action.Kind
}

func (s scriptedStrategy) ScoreAction(_ rules.PlayerObservation, act action.Action) float64 {
	if act.Kind == s.preferKind {
		return 1
	}
	return 0
}

func (scriptedStrategy) ChooseChoice(_ rules.PlayerObservation, request game.ChoiceRequest) []int {
	if len(request.Options) == 0 {
		return nil
	}
	return []int{request.Options[0].Index}
}

func TestAgentChoosesHighestScoringAction(t *testing.T) {
	agent := Agent{Strategy: scriptedStrategy{preferKind: action.ActionCastSpell}}
	// A cast-spell action that is not first, so position alone would not select it.
	legal := []action.Action{action.Pass(), castAction(t)}

	got := agent.ChooseAction(rules.PlayerObservation{}, legal)
	if got.Kind != action.ActionCastSpell {
		t.Errorf("ChooseAction picked %v, want the higher-scoring ActionCastSpell", got.Kind)
	}
}

func TestAgentTieBreaksToFirstLegal(t *testing.T) {
	// BaselineStrategy scores everything equally, so the Agent must keep the
	// engine's ordering and pick the first legal action.
	agent := Agent{Strategy: BaselineStrategy{}}
	first := castAction(t)
	legal := []action.Action{first, action.Pass()}

	if got := agent.ChooseAction(rules.PlayerObservation{}, legal); got.Kind != action.ActionCastSpell {
		t.Errorf("ChooseAction picked %v, want the first legal action", got.Kind)
	}
}

func TestAgentEmptyLegalPasses(t *testing.T) {
	agent := Agent{Strategy: BaselineStrategy{}}
	if got := agent.ChooseAction(rules.PlayerObservation{}, nil); got.Kind != action.ActionPass {
		t.Errorf("ChooseAction on empty legal = %v, want Pass", got.Kind)
	}
}

func TestAgentChooseChoiceDelegatesToStrategy(t *testing.T) {
	agent := Agent{Strategy: scriptedStrategy{}}
	request := game.ChoiceRequest{
		Options:    []game.ChoiceOption{{Index: 0}, {Index: 1}},
		MinChoices: 1,
		MaxChoices: 1,
	}
	got := agent.ChooseChoice(rules.PlayerObservation{}, request)
	if len(got) != 1 || got[0] != 0 {
		t.Errorf("ChooseChoice = %v, want [0] from the strategy", got)
	}
}

func TestBaselineStrategyChooseChoicePrefersDefault(t *testing.T) {
	request := game.ChoiceRequest{
		Options:          []game.ChoiceOption{{Index: 0}, {Index: 1}, {Index: 2}},
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{2},
	}
	got := BaselineStrategy{}.ChooseChoice(rules.PlayerObservation{}, request)
	if len(got) != 1 || got[0] != 2 {
		t.Errorf("ChooseChoice = %v, want the default [2]", got)
	}
}

func TestBaselineStrategyChooseChoiceFirstRequiredOptions(t *testing.T) {
	request := game.ChoiceRequest{
		Options:    []game.ChoiceOption{{Index: 5}, {Index: 6}, {Index: 7}},
		MinChoices: 2,
		MaxChoices: 3,
	}
	got := BaselineStrategy{}.ChooseChoice(rules.PlayerObservation{}, request)
	if len(got) != 2 || got[0] != 5 || got[1] != 6 {
		t.Errorf("ChooseChoice = %v, want the first two options [5 6]", got)
	}
}

func TestAgentPlaysFullGame(t *testing.T) {
	// A land-only game exercises the Agent as a PlayerAgent end to end, driving
	// it to a terminal state. The engine ChoiceAgent path is covered separately
	// by TestAgentAnswersEngineChoices.
	engine := rules.NewEngine(rand.New(rand.NewPCG(1, 2)))

	var configs [game.NumPlayers]game.PlayerConfig
	for i := range configs {
		for range 12 {
			configs[i].Deck = append(configs[i].Deck, forest())
		}
	}

	g := engine.NewGame(configs)
	agents := [game.NumPlayers]rules.PlayerAgent{}
	for i := range agents {
		agents[i] = Agent{Strategy: BaselineStrategy{}}
	}

	result := engine.RunGame(g, agents)
	if result == nil {
		t.Fatal("RunGame returned nil result")
	}
	if result.TurnCount == 0 {
		t.Error("game produced no turns")
	}
	if !g.IsGameOver() {
		t.Error("game did not reach a terminal state")
	}
}

// choiceCountingStrategy delegates to BaselineStrategy but counts how many
// engine-mediated choices it is asked to answer.
type choiceCountingStrategy struct {
	BaselineStrategy

	choices *int
}

func (s choiceCountingStrategy) ChooseChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int {
	*s.choices++
	return s.BaselineStrategy.ChooseChoice(obs, request)
}

func TestAgentAnswersEngineChoices(t *testing.T) {
	// Decks of free Scry sorceries force the engine to raise scry choices, which
	// must be routed through the Agent's ChoiceAgent implementation by RunGame.
	engine := rules.NewEngine(rand.New(rand.NewPCG(1, 2)))

	var configs [game.NumPlayers]game.PlayerConfig
	for i := range configs {
		for range 12 {
			configs[i].Deck = append(configs[i].Deck, scrySpell())
		}
	}

	g := engine.NewGame(configs)
	choiceCount := 0
	agents := [game.NumPlayers]rules.PlayerAgent{}
	for i := range agents {
		agents[i] = Agent{Strategy: choiceCountingStrategy{choices: &choiceCount}}
	}

	engine.RunGame(g, agents)
	if choiceCount == 0 {
		t.Error("the engine never routed a choice through the Agent's ChooseChoice")
	}
}

func forest() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Forest",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}}
}

func scrySpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Free Scry",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{
				{Primitive: game.Scry{Amount: game.Fixed(1), Player: game.ControllerReference()}},
			},
		}.Ability()),
	}}
}

func castAction(t *testing.T) action.Action {
	t.Helper()
	// A minimally valid cast-spell action (card id 1).
	return action.CastSpell(1, nil, 0, nil)
}
