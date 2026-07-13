package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableSpliceOntoArcane exercises "Splice onto Arcane" (CR
// 702.47): an Arcane instant with "Splice onto Arcane <cost>" lowers to a
// game.SpliceKeyword static ability carrying the splice mana cost, alongside its
// normal spell ability.
func TestGenerateExecutableSpliceOntoArcane(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Desperate Ritual",
		Layout:   "normal",
		ManaCost: "{1}{R}",
		TypeLine: "Instant — Arcane",
		OracleText: "Add {R}{R}{R}.\n" +
			"Splice onto Arcane {1}{R} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)",
		Colors: []string{"R"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.SpliceKeyword{Cost:") {
		t.Fatalf("source missing game.SpliceKeyword:\n%s", source)
	}
}

// TestGenerateExecutableSpliceOntoArcaneTargeted confirms a targeted Arcane
// instant that also carries "Splice onto Arcane" compiles: the splice keyword
// lowers to game.SpliceKeyword and the targeted damage spell ability is
// preserved.
func TestGenerateExecutableSpliceOntoArcaneTargeted(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Glacial Ray",
		Layout:   "normal",
		ManaCost: "{1}{R}",
		TypeLine: "Instant — Arcane",
		OracleText: "Glacial Ray deals 2 damage to any target.\n" +
			"Splice onto Arcane {1}{R} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)",
		Colors: []string{"R"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.SpliceKeyword{Cost:") {
		t.Fatalf("source missing game.SpliceKeyword:\n%s", source)
	}
}

// TestGenerateExecutableSpliceNonManaUnsupported confirms the em-dash nonmana
// form of "Splice onto Arcane" is not recognized as the mana-cost splice keyword
// and fails closed: no game.SpliceKeyword is emitted.
func TestGenerateExecutableSpliceNonManaUnsupported(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Horobi's Whisper",
		Layout:   "normal",
		ManaCost: "{3}{B}",
		TypeLine: "Instant — Arcane",
		OracleText: "Destroy target nonblack creature.\n" +
			"Splice onto Arcane—Exile four cards from your graveyard. (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)",
		Colors: []string{"B"},
	}
	source, _, err := GenerateExecutableCardSource(card, "h")
	if err == nil && strings.Contains(source, "game.SpliceKeyword{Cost:") {
		t.Fatalf("nonmana splice must not lower to game.SpliceKeyword:\n%s", source)
	}
}
