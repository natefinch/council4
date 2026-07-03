package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

func containsActionKind(actions []action.Action, kind action.Kind) bool {
	for i := range actions {
		if actions[i].Kind == kind {
			return true
		}
	}
	return false
}

func TestBudgetKeepsTopActionsAndPass(t *testing.T) {
	e := searchTestEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, landCardDef("Forest", mana.G))
	addObservedHandCard(g, game.Player1, creatureWithCost("Bear", 3, 3, 0))
	setAgentMainPhasePriority(g, game.Player1)

	legal := e.Simulator().LegalActions(g, game.Player1)
	searcher := Searcher{Rollout: GenericStrategy{}, Budget: 1}
	candidates := searcher.candidateActions(g, game.Player1, legal)

	if len(candidates) > 2 {
		t.Fatalf("budget 1 kept %d candidates, want at most the top action plus Pass", len(candidates))
	}
	if !containsActionKind(candidates, action.ActionPlayLand) {
		t.Fatalf("budget dropped the play-land, the rollout policy's top-ranked action: %v", candidates)
	}
	if !containsActionKind(candidates, action.ActionPass) {
		t.Fatal("budget did not keep Pass as the do-nothing baseline")
	}
}

func TestSearcherWithBudgetStillDevelopsBoard(t *testing.T) {
	e := searchTestEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, creatureWithCost("Bear", 3, 3, 0))
	setAgentMainPhasePriority(g, game.Player1)

	sim := e.Simulator()
	legal := sim.LegalActions(g, game.Player1)
	searcher := Searcher{Rollout: GenericStrategy{}, Budget: 1}

	if chosen := searcher.searchBestAction(sim, g, game.Player1, legal); chosen.Kind != action.ActionCastSpell {
		t.Fatalf("budgeted searcher chose %v, want to still cast the creature", chosen.Kind)
	}
}
