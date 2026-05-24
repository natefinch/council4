package mana

import "testing"

func TestPoolAmountIncludesSnowAndNonSnowMana(t *testing.T) {
	pool := NewPool()
	pool.Add(Green, 1)
	pool.AddSnow(Green, 2)

	if got := pool.Amount(Green); got != 3 {
		t.Fatalf("green mana = %d, want 3", got)
	}
	if got := pool.SnowAmount(); got != 2 {
		t.Fatalf("snow mana = %d, want 2", got)
	}
}

func TestPoolSpendPreservesSnowManaWhenPossible(t *testing.T) {
	pool := NewPool()
	pool.AddSnow(Green, 1)
	pool.Add(Green, 1)

	if !pool.Spend(Green, 1) {
		t.Fatal("Spend(Green, 1) = false, want true")
	}
	if got := pool.SnowAmount(); got != 1 {
		t.Fatalf("snow mana = %d, want 1 after spending non-snow first", got)
	}
}

func TestPoolSpendSnowRequiresSnowMana(t *testing.T) {
	pool := NewPool()
	pool.Add(Green, 1)

	if pool.SpendSnow(1) {
		t.Fatal("SpendSnow(1) = true with non-snow mana, want false")
	}
	pool.AddSnow(Red, 1)
	if !pool.SpendSnow(1) {
		t.Fatal("SpendSnow(1) = false with snow mana, want true")
	}
	if got := pool.SnowAmount(); got != 0 {
		t.Fatalf("snow mana = %d, want 0", got)
	}
}
