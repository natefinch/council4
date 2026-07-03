package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

func deployScore(t *testing.T, strategy GenericStrategy, ownBoard, opponentBoard int) float64 {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	handID := addObservedHandCard(g, game.Player1, creatureWithCost("Bear", 3, 3, 0))
	for range ownBoard {
		addObservedPermanent(g, game.Player1, creatureCardDef("Mine", 6, 6))
	}
	for range opponentBoard {
		addObservedPermanent(g, game.Player2, creatureCardDef("Theirs", 6, 6))
	}
	obs := rules.NewObservation(g, game.Player1)
	return strategy.ScoreAction(obs, action.CastSpell(handID, nil, 0, nil))
}

func TestArchenemyHoldsBackWhenClearlyAhead(t *testing.T) {
	strategy := GenericStrategy{}
	ahead := deployScore(t, strategy, 2, 0)    // two 6/6s vs an empty table
	atParity := deployScore(t, strategy, 1, 1) // one each

	if ahead >= atParity {
		t.Fatalf("deploying while the clear archenemy (%v) should score below deploying at parity (%v)", ahead, atParity)
	}
}

func TestNoArchenemyPaintWhenNotAhead(t *testing.T) {
	strategy := GenericStrategy{}
	balanced := deployScore(t, strategy, 1, 1)
	empty := deployScore(t, strategy, 0, 0)

	if balanced != empty {
		t.Fatalf("deploying at parity (%v) should not be paint-penalized like the empty board (%v)", balanced, empty)
	}
}

func TestPaintScalePersonality(t *testing.T) {
	if got := (Personality{}).paintScale(); got != 1 {
		t.Fatalf("neutral paintScale = %v, want 1", got)
	}
	if got := (Personality{Aggression: 2}).paintScale(); got >= 1 {
		t.Fatalf("aggressive paintScale = %v, want below 1 (accepts the paint)", got)
	}
	if got := (Personality{PoliticsWeight: 2}).paintScale(); got <= 1 {
		t.Fatalf("political paintScale = %v, want above 1 (avoids the paint)", got)
	}
}
