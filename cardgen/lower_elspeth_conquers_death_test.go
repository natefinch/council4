package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// elspethConquersDeath is the full Saga used by the chapter II and III tests.
func elspethConquersDeath(t *testing.T) loweredFaceAbilities {
	t.Helper()
	return lowerSingleFace(t, &ScryfallCard{
		Name:     "Elspeth Conquers Death",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		ManaCost: "{3}{W}{W}",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I — Exile target permanent an opponent controls with mana value 3 or greater.\n" +
			"II — Noncreature spells your opponents cast cost {2} more to cast until your next turn.\n" +
			"III — Return target creature or planeswalker card from your graveyard to the battlefield. Put a +1/+1 counter or a loyalty counter on it.",
	})
}

// TestLowerElspethChapterTwoNoncreatureTax proves Elspeth Conquers Death's
// chapter II lowers its noncreature-exclusion tax to a resolving,
// duration-bounded cost modifier that excludes creature spells via the negative
// card-type filter.
func TestLowerElspethChapterTwoNoncreatureTax(t *testing.T) {
	t.Parallel()
	face := elspethConquersDeath(t)
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("got %d chapter abilities, want 3", len(face.ChapterAbilities))
	}
	mode := face.ChapterAbilities[1].Content.Modes[0]
	apply := requireApplyRule(t, mode)
	if apply.Duration != game.DurationUntilYourNextTurn {
		t.Fatalf("duration = %v, want DurationUntilYourNextTurn", apply.Duration)
	}
	effect := apply.RuleEffects[0]
	if effect.AffectedPlayer != game.PlayerOpponent {
		t.Fatalf("affected player = %v, want PlayerOpponent", effect.AffectedPlayer)
	}
	if effect.CostModifier.GenericIncrease != 2 {
		t.Fatalf("generic increase = %d, want 2", effect.CostModifier.GenericIncrease)
	}
	if len(effect.CostModifier.CardSelection.RequiredTypes) != 0 {
		t.Fatal("unexpected required card-type filter on chapter II tax")
	}
	if cardSel := effect.CostModifier.CardSelection; len(cardSel.ExcludedTypes) != 1 || cardSel.ExcludedTypes[0] != types.Creature {
		t.Fatalf("excluded card-type filter = %+v, want ExcludedTypes [Creature]", effect.CostModifier.CardSelection)
	}
}

// TestLowerElspethChapterThreeCounterChoice proves Elspeth Conquers Death's
// chapter III lowers its "Return ... to the battlefield. Put a +1/+1 counter or
// a loyalty counter on it." sequence so the reanimated permanent is published
// under a linked key and the trailing counter clause offers the two-kind choice
// on that linked object.
func TestLowerElspethChapterThreeCounterChoice(t *testing.T) {
	t.Parallel()
	face := elspethConquersDeath(t)
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("got %d chapter abilities, want 3", len(face.ChapterAbilities))
	}
	seq := face.ChapterAbilities[2].Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("chapter III sequence length = %d, want 2", len(seq))
	}
	put, ok := seq[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("chapter III primitive[0] = %T, want game.PutOnBattlefield", seq[0].Primitive)
	}
	if put.PublishLinked == "" {
		t.Fatal("PutOnBattlefield.PublishLinked is empty, want a linked key")
	}
	add, ok := seq[1].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("chapter III primitive[1] = %T, want game.AddCounter", seq[1].Primitive)
	}
	if add.Object.Kind() != game.ObjectReferenceLinkedObject {
		t.Fatalf("AddCounter.Object kind = %v, want linked", add.Object.Kind())
	}
	wantKinds := []counter.Kind{counter.PlusOnePlusOne, counter.Loyalty}
	if len(add.KindChoices) != len(wantKinds) {
		t.Fatalf("AddCounter.KindChoices = %v, want %v", add.KindChoices, wantKinds)
	}
	for i, k := range wantKinds {
		if add.KindChoices[i] != k {
			t.Fatalf("AddCounter.KindChoices[%d] = %v, want %v", i, add.KindChoices[i], k)
		}
	}
}
