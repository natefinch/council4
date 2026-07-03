package agent

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

func searchTestEngine() *rules.Engine {
	return rules.NewEngine(rand.New(rand.NewPCG(1, 2)))
}

func setAgentMainPhasePriority(g *game.Game, player game.PlayerID) {
	g.Turn.ActivePlayer = player
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = player
}

func TestSearcherDevelopsBoardOverPassing(t *testing.T) {
	e := searchTestEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, creatureWithCost("Bear", 3, 3, 0))
	setAgentMainPhasePriority(g, game.Player1)

	sim := e.Simulator()
	legal := sim.LegalActions(g, game.Player1)
	searcher := Searcher{Rollout: GenericStrategy{}}

	chosen := searcher.searchBestAction(sim, g, game.Player1, legal)
	if chosen.Kind != action.ActionCastSpell {
		t.Fatalf("searcher chose %v, want to cast the creature (a better position than passing)", chosen.Kind)
	}
}

func TestSearcherPrefersBiggerThreat(t *testing.T) {
	e := searchTestEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, creatureWithCost("Mouse", 1, 1, 0))
	bigID := addObservedHandCard(g, game.Player1, creatureWithCost("Wurm", 7, 7, 0))
	setAgentMainPhasePriority(g, game.Player1)

	sim := e.Simulator()
	legal := sim.LegalActions(g, game.Player1)
	searcher := Searcher{Rollout: GenericStrategy{}}

	chosen := searcher.searchBestAction(sim, g, game.Player1, legal)
	cast, ok := chosen.CastSpellPayload()
	if !ok {
		t.Fatalf("searcher chose %v, want to cast a creature", chosen.Kind)
	}
	if cast.CardID != bigID {
		t.Fatalf("searcher cast %v, want the bigger creature (higher-valued resulting position)", cast.CardID)
	}
}

func TestSearcherFallbackChoosesLegalAction(t *testing.T) {
	// With no legal actions the searcher passes; the fallback ChooseAction path
	// (used for combat and non-search decisions) defers to the rollout policy.
	searcher := Searcher{Rollout: GenericStrategy{}}
	if got := searcher.searchBestAction(rules.Simulator{}, nil, game.Player1, nil); got.Kind != action.ActionPass {
		t.Fatalf("empty legal set produced %v, want pass", got.Kind)
	}
}
