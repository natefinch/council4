package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestCantBeBlockedThisTurnRejectsEveryBlockerUntilCleanup models the runtime
// behavior of Rogue's Passage's activated ability: a temporary
// RuleEffectCantBeBlocked applied to one creature for the turn must make that
// creature unblockable by every legal blocker, leave unrelated attackers
// blockable, and stop applying once the turn's cleanup expires the effect.
func TestCantBeBlockedThisTurnRejectsEveryBlockerUntilCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	otherAttacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	firstBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	secondBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:             game.RuleEffectCantBeBlocked,
		AffectedObjectID: attacker.ObjectID,
		Duration:         game.DurationThisTurn,
		CreatedTurn:      g.Turn.TurnNumber,
	})

	for _, blocker := range []*game.Permanent{firstBlocker, secondBlocker} {
		if canBlockAttacker(g, blocker, attacker) {
			t.Fatal("legal blocker could block creature affected by can't-be-blocked-this-turn effect")
		}
		if !canBlockAttacker(g, blocker, otherAttacker) {
			t.Fatal("can't-be-blocked-this-turn effect prevented blocking an unrelated attacker")
		}
	}

	expireRuleEffects(g)

	for _, blocker := range []*game.Permanent{firstBlocker, secondBlocker} {
		if !canBlockAttacker(g, blocker, attacker) {
			t.Fatal("can't-be-blocked-this-turn effect still applied after cleanup expiry")
		}
	}
}

// TestApplyRuleSourceCantBeBlockedThisTurnGrantAndExpiry drives the exact
// instruction the back-reference lowering produces for "... put a +1/+1 counter
// on this creature. It can't be blocked this turn." (Kappa Cannoneer, Sahagin,
// Razzle-Dazzler): an ApplyRule whose Object is the source permanent, granting a
// RuleEffectCantBeBlocked for the turn. Resolving it must make the source
// unblockable by every legal blocker for the turn, and the grant must lapse once
// the turn's cleanup expires it, restoring normal blocking.
func TestApplyRuleSourceCantBeBlockedThisTurnGrantAndExpiry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID}

	resolveInstruction(engine, g, obj, game.ApplyRule{
		Object: opt.Val(game.SourcePermanentReference()),
		RuleEffects: []game.RuleEffect{
			{Kind: game.RuleEffectCantBeBlocked},
		},
		Duration: game.DurationThisTurn,
	}, nil)

	if canBlockAttacker(g, blocker, source) {
		t.Fatal("source-granted can't-be-blocked-this-turn effect did not make the source unblockable")
	}

	expireRuleEffects(g)

	if !canBlockAttacker(g, blocker, source) {
		t.Fatal("source-granted can't-be-blocked-this-turn effect still applied after cleanup expiry")
	}
}
