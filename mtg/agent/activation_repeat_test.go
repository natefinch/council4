package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

// TestGenericDoesNotRepeatFreeActivation checks that a zero-cost activated
// ability, once activated this turn, is not scored above passing on a repeat.
// Equip has no per-turn limit and X abilities can be activated for X = 0, so a
// free ability scored as a fresh gain each time would be re-activated without end,
// spinning the priority loop (Lightning Greaves re-equipping, Mirror Entity at
// X = 0). The first activation is still a normal play; only the redundant repeat
// is held at or below passing.
func TestGenericDoesNotRepeatFreeActivation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	artifact := addObservedPermanent(g, game.Player1, activatedArtifact("Free Engine", nil))
	act := action.ActivateAbility(artifact.ObjectID, 0, nil, 0)
	strategy := GenericStrategy{}

	if first := strategy.ScoreAction(rules.NewObservation(g, game.Player1), act); first <= scorePass {
		t.Fatalf("first free activation scored %v, want above pass %v", first, scorePass)
	}

	// Record that the ability has already been activated this turn.
	g.AbilityActivationsThisTurn[game.ActivatedAbilityUse{SourceID: artifact.ObjectID, AbilityIndex: 0}] = 1
	if repeat := strategy.ScoreAction(rules.NewObservation(g, game.Player1), act); repeat > scorePass {
		t.Fatalf("repeated free activation scored %v, want at or below pass %v", repeat, scorePass)
	}
}
