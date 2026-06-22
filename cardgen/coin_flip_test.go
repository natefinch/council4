package cardgen

import (
	"strings"
	"testing"
	"unicode"
)

// TestGenerateExecutableCardSourceTavernSwindler exercises the coin-flip win
// branch: "{T}, Pay 3 life: Flip a coin. If you win the flip, you gain 6 life."
// The flip lowers to a fair two-sided RollDie that publishes its result, and the
// gain-life branch is gated on the winning (heads) result.
func TestGenerateExecutableCardSourceTavernSwindler(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Tavern Swindler",
		Layout:     "normal",
		TypeLine:   "Creature — Human Rogue",
		ManaCost:   "{1}{B}",
		Power:      new("1"),
		Toughness:  new("1"),
		OracleText: "{T}, Pay 3 life: Flip a coin. If you win the flip, you gain 6 life.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.RollDie{Sides: 2}",
		`PublishResult: game.ResultKey("coin-flip-result")`,
		"game.GainLife{",
		`ResultGate: opt.Val(game.InstructionResultGate{`,
		`Key:         "coin-flip-result"`,
		"AmountRange: opt.Val(game.IntRange{Min: 2, Max: 2})",
	} {
		if !strings.Contains(normalizeSource(source), normalizeSource(wanted)) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCoinFlipWinLose exercises both branches:
// a win branch gated on heads and a lose branch gated on tails.
func TestGenerateExecutableCardSourceCoinFlipWinLose(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Krark's Other Thumb",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{2}",
		OracleText: "{T}: Flip a coin. If you win the flip, you gain 2 life. If you lose the flip, you lose 2 life.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "k")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.RollDie{Sides: 2}",
		"game.GainLife{",
		"game.LoseLife{",
		"AmountRange: opt.Val(game.IntRange{Min: 2, Max: 2})",
		"AmountRange: opt.Val(game.IntRange{Min: 1, Max: 1})",
	} {
		if !strings.Contains(normalizeSource(source), normalizeSource(wanted)) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// normalizeSource drops every whitespace character and comma so a wanted
// substring matches regardless of the gofmt line breaks, indentation, and
// trailing commas the generated source carries.
func normalizeSource(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == ',' {
			return -1
		}
		return r
	}, s)
}

// TestGenerateExecutableCardSourceCoinFlipTargetedBranchFailsClosed confirms a
// coin flip whose branch targets is not silently emitted ungated: the runtime
// model gates only non-targeted branch effects, so a targeted branch must fail
// closed for the whole ability rather than dropping the flip.
func TestGenerateExecutableCardSourceCoinFlipTargetedBranchFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Crooked Scales",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{4}",
		OracleText: "{4}, {T}: Flip a coin. If you win the flip, destroy target creature an opponent controls. If you lose the flip, destroy target creature you control.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatalf("expected a fail-closed diagnostic, got source:\n%s", source)
	}
	if strings.Contains(normalizeSource(source), normalizeSource("game.RollDie{Sides: 2}")) {
		t.Fatalf("targeted coin-flip branch must not emit a coin flip:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceLoyaltyCoinFlipFailsClosed confirms a coin
// flip on a loyalty ability is not silently emitted ungated. Every ability
// shell must propagate the recognized flip so the lowering either gates it or
// fails closed; a shell that dropped the flip would emit an unconditional
// effect with no RollDie.
func TestGenerateExecutableCardSourceLoyaltyCoinFlipFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Coinflip Walker",
		Layout:     "normal",
		TypeLine:   "Legendary Planeswalker — Test",
		ManaCost:   "{2}{R}",
		Loyalty:    new("3"),
		OracleText: "[+1]: Flip a coin. If you win the flip, you gain 2 life.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatalf("expected a fail-closed diagnostic, got source:\n%s", source)
	}
	if strings.Contains(normalizeSource(source), normalizeSource("game.GainLife{")) {
		t.Fatalf("loyalty coin flip must not emit an ungated effect:\n%s", source)
	}
}
