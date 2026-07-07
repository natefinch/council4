package cardgen

import (
	"strings"
	"testing"
)

// TestLowerMonarchCantAttackYou verifies that a battlefield-scoped "can't attack
// you" restriction gated on the monarch designation ("As long as you're the
// monarch, creatures with power 2 or less can't attack you.", Queen Mother
// Ramonda) lowers onto a conditioned static ability whose Condition carries the
// designation predicate and whose rule effect is the direct-only can't-attack-you
// restriction scoped to the every-creature group narrowed by a power bound. The
// runtime re-evaluates the condition each time rule effects are gathered, so the
// restriction turns on and off as the designation changes.
func TestLowerMonarchCantAttackYou(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Monarch Cant Attack",
		Layout:     "normal",
		ManaCost:   "{3}{W}{W}",
		TypeLine:   "Legendary Creature — Human Noble",
		OracleText: "As long as you're the monarch, creatures with power 2 or less can't attack you.",
		Colors:     []string{"W"},
		Power:      new("3"),
		Toughness:  new("5"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	for _, want := range []string{
		"ControllerIsMonarch: true",
		"Kind:                      game.RuleEffectCantAttack,",
		"DefendingPlayer:           game.PlayerYou,",
		"DefendingPlayerDirectOnly: true,",
		"PermanentTypes:            []types.Card{types.Creature},",
		"AffectedSelection:         game.Selection{Power: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2})},",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestLowerMonarchCantAttackYouRejectsUnconditional proves the battlefield
// "can't attack you" recognizer is gated on the live monarch designation. The
// same restriction without the designation gate ("Creatures with power 2 or less
// can't attack you.", Reverence) has no runtime on/off switch, so it stays
// fail-closed rather than lowering onto an always-on static.
func TestLowerMonarchCantAttackYouRejectsUnconditional(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Unconditional Cant Attack",
		Layout:     "normal",
		ManaCost:   "{2}{W}",
		TypeLine:   "Enchantment",
		OracleText: "Creatures with power 2 or less can't attack you.",
		Colors:     []string{"W"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("unconditional battlefield can't-attack-you unexpectedly lowered")
	}
}
