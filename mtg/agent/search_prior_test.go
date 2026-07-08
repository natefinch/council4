package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func landInHandDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:          name,
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.G)},
	}}
}

// TestSearcherFollowsPolicyPriorAndDevelops checks that on a decision the
// position evaluation cannot separate, the searcher follows the rollout policy's
// prior rather than picking arbitrarily. Playing a land nets about zero to the
// position value — the mana source gained cancels the card spent from hand — so a
// value-only search would not reliably develop its mana. Pass is placed first so
// a value-only tie-break would take it; the rollout policy scores playing a land
// far above passing, and the small policy prior carries that through, so the
// searcher develops.
func TestSearcherFollowsPolicyPriorAndDevelops(t *testing.T) {
	e := searchTestEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, landInHandDef("Forest"))
	setAgentMainPhasePriority(g, game.Player1)

	sim := e.Simulator()
	var playLand action.Action
	for _, a := range sim.LegalActions(g, game.Player1) {
		if a.Kind == action.ActionPlayLand {
			playLand = a
			break
		}
	}
	if playLand.Kind != action.ActionPlayLand {
		t.Fatal("expected playing the land to be a legal action")
	}

	legal := []action.Action{action.Pass(), playLand}
	searcher := Searcher{Rollout: GenericStrategy{}}

	chosen := searcher.searchBestAction(sim, g, game.Player1, legal)
	if chosen.Kind != action.ActionPlayLand {
		t.Fatalf("searcher chose %v, want to develop by playing the land", chosen.Kind)
	}
}
