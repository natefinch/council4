package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestCanAttackAsThoughDefenderRuleEffectPermitsAttack models the runtime
// behavior of the temporary "This creature can attack this turn as though it
// didn't have defender." resolving effect: a RuleEffectCanAttackAsThoughDefender
// scoped to a single affected creature (as the ApplyRule lowering produces) lets
// that Defender creature be declared as an attacker, while another Defender
// creature without the effect remains unable to attack.
func TestCanAttackAsThoughDefenderRuleEffectPermitsAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permitted := addCombatCreaturePermanent(g, game.Player1, game.Defender)
	stuck := addCombatCreaturePermanent(g, game.Player1, game.Defender)

	if canAttackWith(g, permitted, game.Player1) {
		t.Fatal("Defender creature could attack before the rule effect was applied")
	}

	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:               g.IDGen.Next(),
		Kind:             game.RuleEffectCanAttackAsThoughDefender,
		Controller:       game.Player1,
		AffectedObjectID: permitted.ObjectID,
		Duration:         game.DurationThisTurn,
		CreatedTurn:      g.Turn.TurnNumber,
	})

	if !canAttackWith(g, permitted, game.Player1) {
		t.Fatal("can-attack-as-though-defender rule effect did not let the affected Defender attack")
	}
	if canAttackWith(g, stuck, game.Player1) {
		t.Fatal("can-attack-as-though-defender rule effect wrongly let an unaffected Defender attack")
	}
}
