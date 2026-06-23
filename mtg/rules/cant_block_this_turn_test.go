package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestCantBlockThisTurnRuleEffectProhibitsBlocking models the runtime behavior of
// the temporary "<targets> can't block this turn." resolving effect: an
// unconditional RuleEffectCantBlock scoped to a single affected creature (as the
// ApplyRule lowering produces, one per target) stops that creature from blocking
// any attacker, while an unaffected creature blocks normally.
func TestCantBlockThisTurnRuleEffectProhibitsBlocking(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	restricted := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	free := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:               g.IDGen.Next(),
		Kind:             game.RuleEffectCantBlock,
		Controller:       game.Player1,
		AffectedObjectID: restricted.ObjectID,
		Duration:         game.DurationThisTurn,
		CreatedTurn:      g.Turn.TurnNumber,
	})

	if canBlockWith(g, restricted, game.Player2) {
		t.Fatal("can't-block-this-turn rule effect let the affected creature block")
	}
	if !canBlockWith(g, free, game.Player2) {
		t.Fatal("can't-block-this-turn rule effect stopped an unaffected creature from blocking")
	}
}
