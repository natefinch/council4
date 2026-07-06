package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestAttacksDifferentPlayerThanAnother covers Canal Courier's trigger relation:
// it holds only when the source and at least one other attacker attack different
// players.
func TestAttacksDifferentPlayerThanAnother(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCreaturePermanent(g, game.Player1)
	other := addCreaturePermanent(g, game.Player1)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: other.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}}
	event := game.Event{Kind: game.EventAttackerDeclared, AttackTarget: game.AttackTarget{Player: game.Player2}}

	if !attacksDifferentPlayerThanAnother(g, source, event) {
		t.Fatal("want true: source attacks Player2, another attacker attacks Player3")
	}

	// When every other attacker attacks the same player as the source, the
	// relation does not hold.
	g.Combat.Attackers[1].Target.Player = game.Player2
	if attacksDifferentPlayerThanAnother(g, source, event) {
		t.Fatal("want false: both attackers attack Player2")
	}

	// With no other attackers, the relation does not hold.
	g.Combat.Attackers = g.Combat.Attackers[:1]
	if attacksDifferentPlayerThanAnother(g, source, event) {
		t.Fatal("want false: the source is the only attacker")
	}
}

// TestAttacksDifferentPlayerRequiresPlayerAttacks proves that attacks against a
// planeswalker or battle do not count as "attacking a player" for Canal
// Courier's relation, so the trigger does not fire when either the source or the
// other creature is attacking a non-player (per the card ruling).
func TestAttacksDifferentPlayerRequiresPlayerAttacks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCreaturePermanent(g, game.Player1)
	other := addCreaturePermanent(g, game.Player1)

	// The source attacks a planeswalker controlled by Player2 (not a player
	// attack); the other creature attacks Player3 directly. The relation must not
	// hold because the source is not attacking a player.
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: source.ObjectID}},
		{Attacker: other.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}}
	pwEvent := game.Event{Kind: game.EventAttackerDeclared, AttackTarget: g.Combat.Attackers[0].Target}
	if attacksDifferentPlayerThanAnother(g, source, pwEvent) {
		t.Fatal("want false: the source attacks a planeswalker, not a player")
	}

	// The source attacks Player2 directly; the only other attacker attacks a
	// battle. The relation must not hold because the other creature is not
	// attacking a player.
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: other.ObjectID, Target: game.AttackTarget{Player: game.Player3, BattleID: other.ObjectID}},
	}}
	playerEvent := game.Event{Kind: game.EventAttackerDeclared, AttackTarget: g.Combat.Attackers[0].Target}
	if attacksDifferentPlayerThanAnother(g, source, playerEvent) {
		t.Fatal("want false: the other attacker attacks a battle, not a player")
	}
}

// TestExpireEndOfCombatRuleEffects proves that "this combat" rule effects
// (DurationUntilEndOfCombat) are removed when combat is torn down, while other
// durations survive.
func TestExpireEndOfCombatRuleEffects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = []game.RuleEffect{
		{Kind: game.RuleEffectCantBeBlocked, Duration: game.DurationUntilEndOfCombat},
		{Kind: game.RuleEffectCantBeBlocked, Duration: game.DurationThisTurn},
	}

	expireEndOfCombatRuleEffects(g)

	if len(g.RuleEffects) != 1 {
		t.Fatalf("rule effects remaining = %d, want 1 (only the this-turn effect)", len(g.RuleEffects))
	}
	if g.RuleEffects[0].Duration != game.DurationThisTurn {
		t.Fatalf("remaining effect duration = %v, want DurationThisTurn", g.RuleEffects[0].Duration)
	}
}
