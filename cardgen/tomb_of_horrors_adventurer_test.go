package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerTombOfHorrorsAdventurerConditionalCopies(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Tomb of Horrors Adventurer",
		Layout:   "normal",
		TypeLine: "Creature — Elf Monk",
		OracleText: "When this creature enters, you take the initiative.\n" +
			"Whenever you cast your second spell each turn, copy it. " +
			"If you've completed a dungeon, copy that spell twice instead. " +
			"You may choose new targets for the copies. " +
			"(A copy of a permanent spell becomes a token.)",
	})
	var copyTrigger *game.TriggeredAbility
	for i := range face.TriggeredAbilities {
		trigger := &face.TriggeredAbilities[i]
		if trigger.Trigger.Pattern.Event == game.EventSpellCast {
			copyTrigger = trigger
			break
		}
	}
	if copyTrigger == nil {
		t.Fatal("spell-cast copy trigger not lowered")
	}
	if copyTrigger.Trigger.Pattern.PlayerEventOrdinalThisTurn != 2 {
		t.Fatalf("spell ordinal = %d, want 2", copyTrigger.Trigger.Pattern.PlayerEventOrdinalThisTurn)
	}
	sequence := copyTrigger.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v", sequence)
	}
	for i, wantCount := range []int{1, 2} {
		copyEffect, ok := sequence[i].Primitive.(game.CopyStackObject)
		if !ok ||
			copyEffect.Object != game.EventStackObjectReference() ||
			copyEffect.Count != wantCount ||
			!copyEffect.MayChooseNewTargets {
			t.Fatalf("copy instruction %d = %#v", i, sequence[i])
		}
		if !sequence[i].Condition.Exists || !sequence[i].Condition.Val.Condition.Exists {
			t.Fatalf("copy instruction %d missing dungeon gate", i)
		}
		condition := sequence[i].Condition.Val.Condition.Val
		if !condition.ControllerCompletedADungeon || condition.Negate != (i == 0) {
			t.Fatalf("copy instruction %d condition = %#v", i, condition)
		}
	}
}
