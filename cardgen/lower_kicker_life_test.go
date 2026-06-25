package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerKickedConditionalGainOrLoseLife proves the marquee shape of
// Sheoldred's Restoration: a reanimation followed by a kicked-conditional
// gain-or-lose life rider. "Return target creature card from your graveyard to
// the battlefield. If this spell was kicked, you gain life equal to that card's
// mana value. Otherwise, you lose that much life." lowers to:
//
//	seq0: PutOnBattlefield publishing the entered permanent under a linked key
//	seq1: GainLife of that permanent's mana value, gated on SpellWasKicked and on
//	      the move reaching the battlefield
//	seq2: LoseLife of the same amount, gated on NOT SpellWasKicked and the same
//	      move result
//
// Both branches read the identical linked mana-value amount, so the "Otherwise"
// loss mirrors the gain the kicked spell would have produced.
func TestLowerKickedConditionalGainOrLoseLife(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Sheoldred's Restoration",
		Layout:   "normal",
		ManaCost: "{3}{B}",
		TypeLine: "Sorcery",
		OracleText: "Kicker {2}{W}\n" +
			"Return target creature card from your graveyard to the battlefield. " +
			"If this spell was kicked, you gain life equal to that card's mana value. " +
			"Otherwise, you lose that much life.\n" +
			"Exile Sheoldred's Restoration.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 4 {
		t.Fatalf("mode = %+v, want one target and four instructions", mode)
	}

	move, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok {
		t.Fatalf("seq0 = %T, want game.PutOnBattlefield", mode.Sequence[0].Primitive)
	}
	if move.PublishLinked == "" {
		t.Fatalf("reanimation = %+v, want a published linked key", move)
	}

	gain, ok := mode.Sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("seq1 = %T, want game.GainLife", mode.Sequence[1].Primitive)
	}
	assertSpellKickedGate(t, mode.Sequence[1], false)
	assertMoveResultGate(t, mode.Sequence[1])
	gainDyn := gain.Amount.DynamicAmount()
	if !gainDyn.Exists || gainDyn.Val.Kind != game.DynamicAmountObjectManaValue {
		t.Fatalf("gain amount = %+v, want ObjectManaValue dynamic", gain.Amount)
	}
	if gainDyn.Val.Object != game.LinkedObjectReference(string(move.PublishLinked)) {
		t.Fatalf("gain object = %+v, want linked permanent %q", gainDyn.Val.Object, move.PublishLinked)
	}

	lose, ok := mode.Sequence[2].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("seq2 = %T, want game.LoseLife", mode.Sequence[2].Primitive)
	}
	assertSpellKickedGate(t, mode.Sequence[2], true)
	assertMoveResultGate(t, mode.Sequence[2])
	if lose.Amount != gain.Amount {
		t.Fatalf("lose amount = %+v, want identical to gain amount %+v", lose.Amount, gain.Amount)
	}
	if lose.Player != game.ControllerReference() {
		t.Fatalf("lose player = %+v, want controller", lose.Player)
	}

	if _, ok := mode.Sequence[3].Primitive.(game.Exile); !ok {
		t.Fatalf("seq3 = %T, want game.Exile (self-exile)", mode.Sequence[3].Primitive)
	}

	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1 (Kicker)", len(face.StaticAbilities))
	}
	if _, ok := game.BodyKeywordAbility(&face.StaticAbilities[0].Body, game.Kicker); !ok {
		t.Fatalf("Kicker keyword not found in %#v", face.StaticAbilities[0].Body)
	}
}

func assertSpellKickedGate(t *testing.T, instruction game.Instruction, negated bool) {
	t.Helper()
	if !instruction.Condition.Exists || !instruction.Condition.Val.Condition.Exists {
		t.Fatalf("instruction %+v has no kicked condition gate", instruction)
	}
	cond := instruction.Condition.Val.Condition.Val
	if !cond.SpellWasKicked {
		t.Fatalf("gate = %#v, want SpellWasKicked", cond)
	}
	if cond.Negate != negated {
		t.Fatalf("gate negate = %v, want %v", cond.Negate, negated)
	}
}

func assertMoveResultGate(t *testing.T, instruction game.Instruction) {
	t.Helper()
	if !instruction.ResultGate.Exists {
		t.Fatalf("instruction %+v has no move result gate", instruction)
	}
	if instruction.ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("result gate = %+v, want Succeeded=TriTrue", instruction.ResultGate.Val)
	}
}
