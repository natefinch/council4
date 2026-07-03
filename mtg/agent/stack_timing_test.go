package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/rules"
)

func removalScoreOnTurn(active game.PlayerID, target *game.CardDef) float64 {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	boltID := addObservedHandCard(g, game.Player1, instantDef("Zap", 1, color.Red))
	enemy := addObservedPermanent(g, game.Player2, target)
	g.Turn.ActivePlayer = active
	obs := rules.NewObservation(g, game.Player1)
	return GenericStrategy{}.ScoreAction(obs, action.CastSpell(boltID,
		[]game.Target{game.PermanentTarget(enemy.ObjectID)}, 0, nil))
}

func TestHoldsInstantRemovalOnOwnTurn(t *testing.T) {
	// A 6/6 is worth removing, but on the agent's own turn the answer should be
	// held for an opponent's turn, scoring below firing it on their turn.
	target := creatureCardDef("Wurm", 6, 6)
	myTurn := removalScoreOnTurn(game.Player1, target)
	theirTurn := removalScoreOnTurn(game.Player2, target)
	if myTurn >= theirTurn {
		t.Fatalf("removal on my turn (%v) should score below removal on an opponent's turn (%v)", myTurn, theirTurn)
	}
	if theirTurn <= scorePass {
		t.Fatalf("removal of a 6/6 on an opponent's turn scored %v, want above pass", theirTurn)
	}
}

func TestRemovesHugeThreatEvenOnOwnTurn(t *testing.T) {
	// A dangerous threat is removed immediately even on the agent's own turn: the
	// hold penalty is modest and does not stop the agent answering a real bomb.
	huge := creatureCardDef("Leviathan", 12, 12)
	if score := removalScoreOnTurn(game.Player1, huge); score <= scorePass {
		t.Fatalf("removing a 12/12 on my own turn scored %v, want above pass", score)
	}
}
