package mana

import "testing"

// TestPoolTracksCreatureProvenance proves a pool reports how much of its mana was
// produced by creatures independently of color, so "mana from creatures" counting
// (Inga and Esika) reads the provenance stamped on each unit.
func TestPoolTracksCreatureProvenance(t *testing.T) {
	pool := NewPool()
	pool.AddUnit(Unit{Color: G, FromCreature: true}, 2)
	pool.AddUnit(Unit{Color: U, FromCreature: true}, 1)
	pool.Add(R, 3)
	if got := pool.CreatureAmount(); got != 3 {
		t.Fatalf("creature mana = %d, want 3", got)
	}
	if got := pool.Total(); got != 6 {
		t.Fatalf("total mana = %d, want 6", got)
	}
}

// TestPoolSpendPreservesCreatureManaWhenPossible proves a plain colored spend
// consumes non-creature mana before creature mana, so a payment that could be
// made either way keeps creature provenance available for later counting.
func TestPoolSpendPreservesCreatureManaWhenPossible(t *testing.T) {
	pool := NewPool()
	pool.AddUnit(Unit{Color: G, FromCreature: true}, 1)
	pool.Add(G, 1)
	if !pool.Spend(G, 1) {
		t.Fatal("Spend(G, 1) = false, want true")
	}
	if got := pool.CreatureAmount(); got != 1 {
		t.Fatalf("creature mana = %d, want 1 after spending non-creature first", got)
	}
}

// TestPoolCreatureAndSnowProvenanceAreIndependent proves creature and snow
// provenance coexist on distinct units without one masking the other.
func TestPoolCreatureAndSnowProvenanceAreIndependent(t *testing.T) {
	pool := NewPool()
	pool.AddUnit(Unit{Color: G, Snow: true, FromCreature: true}, 1)
	pool.AddSnow(G, 1)
	pool.AddUnit(Unit{Color: G, FromCreature: true}, 1)
	if got := pool.SnowAmount(); got != 2 {
		t.Fatalf("snow mana = %d, want 2", got)
	}
	if got := pool.CreatureAmount(); got != 2 {
		t.Fatalf("creature mana = %d, want 2", got)
	}
}
