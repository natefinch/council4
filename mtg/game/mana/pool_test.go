package mana

import (
	"testing"
)

func TestPoolAmountIncludesSnowAndNonSnowMana(t *testing.T) {
	pool := NewPool()
	pool.Add(G, 1)
	pool.AddSnow(G, 2)

	if got := pool.Amount(G); got != 3 {
		t.Fatalf("green mana = %d, want 3", got)
	}
	if got := pool.SnowAmount(); got != 2 {
		t.Fatalf("snow mana = %d, want 2", got)
	}
}

func TestPoolSpendPreservesSnowManaWhenPossible(t *testing.T) {
	pool := NewPool()
	pool.AddSnow(G, 1)
	pool.Add(G, 1)

	if !pool.Spend(G, 1) {
		t.Fatal("Spend(Green, 1) = false, want true")
	}
	if got := pool.SnowAmount(); got != 1 {
		t.Fatalf("snow mana = %d, want 1 after spending non-snow first", got)
	}
}

func TestPoolSpendSnowRequiresSnowMana(t *testing.T) {
	pool := NewPool()
	pool.Add(G, 1)

	if pool.SpendSnow(1) {
		t.Fatal("SpendSnow(1) = true with non-snow mana, want false")
	}
	pool.AddSnow(R, 1)
	if !pool.SpendSnow(1) {
		t.Fatal("SpendSnow(1) = false with snow mana, want true")
	}
	if got := pool.SnowAmount(); got != 0 {
		t.Fatalf("snow mana = %d, want 0", got)
	}
}

func TestPoolTracksSpentTotal(t *testing.T) {
	pool := NewPool()
	pool.Add(G, 2)
	pool.Add(U, 1)

	if got := pool.Spent(); got != 0 {
		t.Fatalf("Spent() = %d, want 0 before spending", got)
	}
	if !pool.Spend(G, 2) {
		t.Fatal("Spend(Green, 2) = false, want true")
	}
	if !pool.Spend(U, 1) {
		t.Fatal("Spend(Blue, 1) = false, want true")
	}
	if got := pool.Spent(); got != 3 {
		t.Fatalf("Spent() = %d, want 3 after spending 3 mana", got)
	}
}

func TestPoolSpentSurvivesEmptyAndCounts(t *testing.T) {
	pool := NewPool()
	pool.Add(R, 3)
	if !pool.Spend(R, 1) {
		t.Fatal("Spend(Red, 1) = false, want true")
	}
	// Failed spends do not count; unspent mana emptied at end of step does not count.
	if pool.Spend(R, 5) {
		t.Fatal("Spend(Red, 5) = true with only 2 left, want false")
	}
	pool.Empty()
	if got := pool.Spent(); got != 1 {
		t.Fatalf("Spent() = %d, want 1 (failed spend and emptied mana excluded)", got)
	}
}

func TestPoolCloneCopiesSpent(t *testing.T) {
	pool := NewPool()
	pool.Add(B, 1)
	pool.Spend(B, 1)

	clone := pool.Clone()
	if got := clone.Spent(); got != 1 {
		t.Fatalf("clone Spent() = %d, want 1", got)
	}
	clone.Add(B, 1)
	clone.Spend(B, 1)
	if got := pool.Spent(); got != 1 {
		t.Fatalf("original Spent() = %d, want 1 (clone spend must not affect original)", got)
	}
	if got := clone.Spent(); got != 2 {
		t.Fatalf("clone Spent() = %d, want 2", got)
	}
}

func TestPoolPersistentManaSurvivesEmptyUntilCleared(t *testing.T) {
	pool := NewPool()
	pool.Add(R, 2)           // ordinary mana
	pool.AddPersistent(R, 3) // persistent mana

	if got := pool.Amount(R); got != 5 {
		t.Fatalf("red mana after adds = %d, want 5", got)
	}

	// A step or phase ending empties the pool; persistent mana is preserved.
	pool.Empty()
	if got := pool.Amount(R); got != 3 {
		t.Fatalf("red mana after Empty = %d, want 3 (only persistent survives)", got)
	}

	// Persistent mana survives repeated boundaries within the turn.
	pool.Empty()
	if got := pool.Amount(R); got != 3 {
		t.Fatalf("red mana after second Empty = %d, want 3", got)
	}

	// End-of-turn cleanup releases the reservation, so the next Empty removes it.
	pool.ClearPersistent()
	pool.Empty()
	if got := pool.Total(); got != 0 {
		t.Fatalf("total mana after ClearPersistent + Empty = %d, want 0", got)
	}
}

func TestPoolSpentPersistentManaDoesNotResurrectOnEmpty(t *testing.T) {
	pool := NewPool()
	pool.AddPersistent(R, 3)

	if !pool.Spend(R, 2) {
		t.Fatal("Spend(R, 2) = false, want true")
	}
	if got := pool.Amount(R); got != 1 {
		t.Fatalf("red mana after spending 2 = %d, want 1", got)
	}

	// The single remaining unit is still persistent; the two spent units must not
	// reappear when the pool empties.
	pool.Empty()
	if got := pool.Amount(R); got != 1 {
		t.Fatalf("red mana after Empty = %d, want 1 (spent persistent mana must not resurrect)", got)
	}
}
