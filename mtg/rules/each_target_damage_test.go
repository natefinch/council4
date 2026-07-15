package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// eachTargetDamage builds an EachTarget damage instruction dealing the full
// fixed amount to every target chosen for the spec at index 0.
func eachTargetDamage(amount int) game.Damage {
	return game.Damage{
		Amount:     game.Fixed(amount),
		Recipient:  game.AnyTargetDamageRecipient(0),
		EachTarget: true,
	}
}

// TestEachTargetDamageDealsFullAmountToEveryTarget proves the "deals X damage to
// each of them" resolution (Comet Storm) deals the whole amount independently to
// every chosen target — two creatures and a player — without splitting it.
func TestEachTargetDamageDealsFullAmountToEveryTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player2)
	second := addCreaturePermanent(g, game.Player2)
	startingLife := g.Players[game.Player2].Life
	addEffectSpellToStack(g, game.Player1, eachTargetDamage(3), []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
		game.PlayerTarget(game.Player2),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	gotFirst, ok := permanentByObjectID(g, first.ObjectID)
	if !ok {
		t.Fatal("first target left the battlefield")
	}
	gotSecond, ok := permanentByObjectID(g, second.ObjectID)
	if !ok {
		t.Fatal("second target left the battlefield")
	}
	if gotFirst.MarkedDamage != 3 {
		t.Fatalf("first target marked damage = %d, want 3", gotFirst.MarkedDamage)
	}
	if gotSecond.MarkedDamage != 3 {
		t.Fatalf("second target marked damage = %d, want 3", gotSecond.MarkedDamage)
	}
	if got := g.Players[game.Player2].Life; got != startingLife-3 {
		t.Fatalf("player life = %d, want %d", got, startingLife-3)
	}
}

// TestEachTargetDamageSingleTarget proves the unkicked shape (one chosen target)
// deals the full amount to that single target.
func TestEachTargetDamageSingleTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	only := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, eachTargetDamage(5), []game.Target{
		game.PermanentTarget(only.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	got, ok := permanentByObjectID(g, only.ObjectID)
	if !ok {
		t.Fatal("target left the battlefield")
	}
	if got.MarkedDamage != 5 {
		t.Fatalf("target marked damage = %d, want 5", got.MarkedDamage)
	}
}

// TestEachTargetDamageSkipsIllegalTarget proves a target that became illegal
// since announcement is skipped at resolution while every still-legal target
// still takes the full amount (CR 608.2b) — the amount is not split, so the
// survivors are unaffected by the illegal target's removal.
func TestEachTargetDamageSkipsIllegalTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player2)
	second := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, eachTargetDamage(4), []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})
	// The second target becomes illegal (e.g. sacrificed) before resolution.
	if _, ok := removePermanentFromBattlefield(g, second.ObjectID); !ok {
		t.Fatal("failed to remove the second target before resolution")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	gotFirst, ok := permanentByObjectID(g, first.ObjectID)
	if !ok {
		t.Fatal("first (still-legal) target left the battlefield")
	}
	if gotFirst.MarkedDamage != 4 {
		t.Fatalf("surviving target marked damage = %d, want the full 4", gotFirst.MarkedDamage)
	}
	if _, stillThere := permanentByObjectID(g, second.ObjectID); stillThere {
		t.Fatal("removed target unexpectedly still on the battlefield")
	}
}
