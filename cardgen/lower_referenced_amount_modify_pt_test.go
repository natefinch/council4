package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// modifyPTFromContent returns the lone ModifyPT primitive of an ability content's
// single mode, failing the test if the content is not exactly one mode with one
// ModifyPT instruction.
func modifyPTFromContent(t *testing.T, content game.AbilityContent) game.ModifyPT {
	t.Helper()
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(content.Modes))
	}
	sequence := content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %d, want 1", len(sequence))
	}
	modify, ok := sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("primitive = %T, want game.ModifyPT", sequence[0].Primitive)
	}
	return modify
}

// TestLowerReferencedAmountTargetPumpSourceCounters proves the referenced-amount
// target pump lowers a single-target until-end-of-turn power/toughness boost whose
// magnitude counts counters on the source permanent ("Target creature gets +X/+X
// until end of turn, where X is the number of verse counters on this
// enchantment.", War Dance). Both deltas resolve to the source's counter count.
func TestLowerReferencedAmountTargetPumpSourceCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Verse Pump",
		Layout:   "normal",
		TypeLine: "Enchantment",
		OracleText: "At the beginning of your upkeep, you may put a verse counter on this enchantment.\n" +
			"Sacrifice this enchantment: Target creature gets +X/+X until end of turn, " +
			"where X is the number of verse counters on this enchantment.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	modify := modifyPTFromContent(t, face.ActivatedAbilities[0].Content)
	if modify.Object != game.TargetPermanentReference(0) {
		t.Fatalf("ModifyPT.Object = %v, want TargetPermanentReference(0)", modify.Object)
	}
	if modify.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("ModifyPT.Duration = %v, want DurationUntilEndOfTurn", modify.Duration)
	}
	for label, delta := range map[string]game.Quantity{"power": modify.PowerDelta, "toughness": modify.ToughnessDelta} {
		amount := delta.DynamicAmount()
		if !amount.Exists {
			t.Fatalf("%s delta = %v, want dynamic", label, delta)
		}
		if amount.Val.Kind != game.DynamicAmountObjectCounters {
			t.Fatalf("%s delta kind = %v, want DynamicAmountObjectCounters", label, amount.Val.Kind)
		}
		if amount.Val.CounterKind != counter.Verse {
			t.Fatalf("%s delta counter = %v, want Verse", label, amount.Val.CounterKind)
		}
		if amount.Val.Object != game.SourcePermanentReference() {
			t.Fatalf("%s delta object = %v, want SourcePermanentReference", label, amount.Val.Object)
		}
	}
}

// TestLowerReferencedAmountTargetPumpTargetManaValue proves the referenced-amount
// target pump lowers a single-target until-end-of-turn boost whose magnitude is
// the pumped creature's own characteristic, named by the pronoun "its" ("Target
// creature gets +0/+X until end of turn, where X is its mana value.", Great
// Defender). The dynamic toughness delta anchors on the target slot itself.
func TestLowerReferencedAmountTargetPumpTargetManaValue(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mana Value Pump",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +0/+X until end of turn, where X is its mana value.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	content := face.SpellAbility.Val
	modify := modifyPTFromContent(t, content)
	if modify.Object != game.TargetPermanentReference(0) {
		t.Fatalf("ModifyPT.Object = %v, want TargetPermanentReference(0)", modify.Object)
	}
	if modify.PowerDelta.IsDynamic() || modify.PowerDelta.Value() != 0 {
		t.Fatalf("power delta = %v, want Fixed(0)", modify.PowerDelta)
	}
	amount := modify.ToughnessDelta.DynamicAmount()
	if !amount.Exists {
		t.Fatalf("toughness delta = %v, want dynamic", modify.ToughnessDelta)
	}
	if amount.Val.Kind != game.DynamicAmountObjectManaValue {
		t.Fatalf("toughness delta kind = %v, want DynamicAmountObjectManaValue", amount.Val.Kind)
	}
	if amount.Val.Object != game.TargetPermanentReference(0) {
		t.Fatalf("toughness delta object = %v, want TargetPermanentReference(0)", amount.Val.Object)
	}
}

// TestLowerReferencedAmountTargetPumpRejectsDemonstrativeReferent proves the
// referenced-amount target pump fails closed when the dynamic amount's referent is
// a demonstrative naming a different object the antecedent binder routed to the
// target slot ("… where X is that spell's mana value.", Livaan, Cultist of
// Tiamat). Reading the target creature's mana value there would silently compute
// the wrong magnitude, so the lowering must reject rather than mis-resolve it.
func TestLowerReferencedAmountTargetPumpRejectsDemonstrativeReferent(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Test Demonstrative Referent",
		Layout:   "normal",
		TypeLine: "Creature — Dragon",
		OracleText: "Whenever you cast a noncreature spell, target creature gets +X/+0 until end of turn, " +
			"where X is that spell's mana value.",
		Power:     new("4"),
		Toughness: new("4"),
	})
}
