package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerEdictThenLoseLifeShareTargetPlayer verifies the edict-plus-payoff
// pattern "Target player sacrifices a creature of their choice and loses N life."
// (Geth's Verdict) lowers as an ordered sequence sharing the one player target:
// the sacrifice and the life loss both address that player. This exercises the
// SacrificePermanents player-target remapping the sequence walker previously
// lacked.
func TestLowerEdictThenLoseLifeShareTargetPlayer(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Verdict",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target player sacrifices a creature of their choice and loses 1 life.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want one shared player target", len(mode.Targets))
	}
	var sacrifice *game.SacrificePermanents
	var lose *game.LoseLife
	for i := range mode.Sequence {
		if prim, ok := mode.Sequence[i].Primitive.(game.SacrificePermanents); ok {
			sacrifice = &prim
		}
		if prim, ok := mode.Sequence[i].Primitive.(game.LoseLife); ok {
			lose = &prim
		}
	}
	if sacrifice == nil || lose == nil {
		t.Fatalf("sequence missing sacrifice or lose-life: %+v", mode.Sequence)
	}
	if sacrifice.Player != game.TargetPlayerReference(0) {
		t.Fatalf("sacrifice player = %+v, want target 0", sacrifice.Player)
	}
	if lose.Player != game.TargetPlayerReference(0) {
		t.Fatalf("lose-life player = %+v, want the same shared target 0", lose.Player)
	}
}

// TestEdictThenGainToughnessFailsClosed verifies the edict payoff that scales by
// the sacrificed creature's characteristics ("Target opponent sacrifices a
// creature of their choice. You gain life equal to that creature's toughness." —
// Tribute to Hunger) fails closed rather than emitting a card whose life gain
// silently resolves to zero: the sacrificed creature is not a target, so its
// toughness reference cannot resolve.
func TestEdictThenGainToughnessFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Tribute",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target opponent sacrifices a creature of their choice. You gain life equal to that creature's toughness.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected a diagnostic for the unsupported sacrificed-creature payoff reference")
	}
}
