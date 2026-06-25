package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestDefendingPlayerGetsPoisonCounterOnUnblockedAttack covers the broadened
// gain-player-counter recipient: a creature with an unblocked-attacker trigger
// whose effect is "defending player gets a poison counter" resolves the
// DefendingPlayer reference off the EventAttackerBecameUnblocked event and adds
// the counter to the defending player, not the attacking controller.
func TestDefendingPlayerGetsPoisonCounterOnUnblockedAttack(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:       game.EventAttackerBecameUnblocked,
			Controller: game.Player1,
			Player:     game.Player2,
		},
	}

	resolveInstruction(engine, g, obj, game.AddPlayerCounter{
		Amount:      game.Fixed(1),
		Player:      game.DefendingPlayerReference(),
		CounterKind: counter.Poison,
	}, &TurnLog{})

	if got := g.Players[game.Player2].PoisonCounters; got != 1 {
		t.Fatalf("defending player poison counters = %d, want 1", got)
	}
	if got := g.Players[game.Player1].PoisonCounters; got != 0 {
		t.Fatalf("attacking controller poison counters = %d, want 0", got)
	}
}

// TestDefendingPlayerReferenceResolvesOnBecameBlocked covers the Afflict-style
// becomes-blocked trigger: the DefendingPlayer reference now resolves off the
// EventAttackerBecameBlocked event's Player subject, so a "defending player ..."
// effect on a becomes-blocked trigger hits the defending player at runtime.
func TestDefendingPlayerReferenceResolvesOnBecameBlocked(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:       game.EventAttackerBecameBlocked,
			Controller: game.Player1,
			Player:     game.Player2,
		},
	}

	resolveInstruction(engine, g, obj, game.AddPlayerCounter{
		Amount:      game.Fixed(2),
		Player:      game.DefendingPlayerReference(),
		CounterKind: counter.Poison,
	}, &TurnLog{})

	if got := g.Players[game.Player2].PoisonCounters; got != 2 {
		t.Fatalf("defending player poison counters = %d, want 2", got)
	}
}

// TestAttackerDefendingPlayerUsesDeclaredTarget confirms the combat engine
// derives an attacker's defending player from its declaration's AttackTarget,
// which is what feeds the Player subject on the became-blocked / became-unblocked
// events.
func TestAttackerDefendingPlayerUsesDeclaredTarget(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, vanillaCreature("Lone Raider", 2, 2))
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	if got := attackerDefendingPlayer(g, attacker.ObjectID); got != game.Player2 {
		t.Fatalf("defending player = %v, want Player2", got)
	}
}
