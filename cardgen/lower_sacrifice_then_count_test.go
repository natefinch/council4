package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerSacrificeThenCountCreateToken proves the anchor card Hellion
// Eruption: "Sacrifice all creatures you control, then create that many 4/4 red
// Hellion creature tokens." The sacrifice publishes the number sacrificed and
// the token creation reads it through DynamicAmountPreviousEffectResult so the
// reward scales to exactly the number of creatures sacrificed.
func TestLowerSacrificeThenCountCreateToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Hellion Eruption",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{5}{R}",
		OracleText: "Sacrifice all creatures you control, then create that many 4/4 red Hellion creature tokens.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2 (sacrifice, create)", len(mode.Sequence))
	}
	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("first primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if !sacrifice.All {
		t.Fatal("sacrifice is not All")
	}
	if sacrifice.AnyNumber {
		t.Fatal("sacrifice must not be AnyNumber for an all-creatures clause")
	}
	if mode.Sequence[0].PublishResult != sacrificedThisWayResultKey {
		t.Fatalf("sacrifice PublishResult = %q, want %q", mode.Sequence[0].PublishResult, sacrificedThisWayResultKey)
	}
	if len(sacrifice.Selection.RequiredTypes) != 1 || sacrifice.Selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("sacrifice selection = %+v, want single Creature type", sacrifice.Selection)
	}
	create, ok := mode.Sequence[1].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("second primitive = %T, want game.CreateToken", mode.Sequence[1].Primitive)
	}
	assertScaledToSacrificedCount(t, "create", create.Amount)
}

// TestLowerSacrificeAnyNumberThenCountAddMana proves the anchor card Mana Seism:
// "Sacrifice any number of lands, then add that much {C}." The any-number
// sacrifice publishes its count and the mana production reads it so the
// controller adds exactly one colorless mana per land sacrificed.
func TestLowerSacrificeAnyNumberThenCountAddMana(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mana Seism",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{1}{R}",
		OracleText: "Sacrifice any number of lands, then add that much {C}.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2 (sacrifice, add mana)", len(mode.Sequence))
	}
	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("first primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if !sacrifice.AnyNumber {
		t.Fatal("sacrifice is not AnyNumber")
	}
	if sacrifice.All {
		t.Fatal("sacrifice must not be All for an any-number clause")
	}
	if len(sacrifice.Selection.RequiredTypes) != 1 || sacrifice.Selection.RequiredTypes[0] != types.Land {
		t.Fatalf("sacrifice selection = %+v, want single Land type", sacrifice.Selection)
	}
	addMana, ok := mode.Sequence[1].Primitive.(game.AddMana)
	if !ok {
		t.Fatalf("second primitive = %T, want game.AddMana", mode.Sequence[1].Primitive)
	}
	if addMana.ManaColor != mana.C {
		t.Fatalf("add mana color = %q, want colorless", addMana.ManaColor)
	}
	assertScaledToSacrificedCount(t, "add mana", addMana.Amount)
}

// TestLowerSacrificeThenCountRejectsTypeUnionSelection proves a count-scaled
// sacrifice over a type-union-or-token pool (Malevolent Witchkite: "sacrifice
// any number of artifacts, enchantments, and/or tokens, then draw that many
// cards.") fails closed: the runtime selection cannot express the token
// disjunct, so the faithful single-type gate leaves the sequence unsupported
// rather than sacrificing the wrong pool.
func TestLowerSacrificeThenCountRejectsTypeUnionSelection(t *testing.T) {
	t.Parallel()
	power, toughness := "4", "4"
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:      "Test Malevolent Witchkite",
		Layout:    "normal",
		TypeLine:  "Creature — Dragon",
		ManaCost:  "{4}{B}{B}",
		Power:     &power,
		Toughness: &toughness,
		OracleText: "Flying\nWhen this creature enters, sacrifice any number of artifacts, " +
			"enchantments, and/or tokens, then draw that many cards.",
	})
	if len(face.TriggeredAbilities) != 0 {
		t.Fatalf("triggered abilities = %d, want 0 (unsupported)", len(face.TriggeredAbilities))
	}
}

func assertScaledToSacrificedCount(t *testing.T, label string, amount game.Quantity) {
	t.Helper()
	dynamic := amount.DynamicAmount()
	if !dynamic.Exists {
		t.Fatalf("%s amount = %+v, want dynamic", label, amount)
	}
	if dynamic.Val.Kind != game.DynamicAmountPreviousEffectResult {
		t.Fatalf("%s dynamic kind = %v, want DynamicAmountPreviousEffectResult", label, dynamic.Val.Kind)
	}
	if dynamic.Val.ResultKey != sacrificedThisWayResultKey {
		t.Fatalf("%s dynamic result key = %q, want %q", label, dynamic.Val.ResultKey, sacrificedThisWayResultKey)
	}
}
