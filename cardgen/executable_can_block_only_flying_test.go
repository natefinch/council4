package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceSelfCanBlockOnlyFlying confirms that the
// blocker-side permission restriction "This creature can block only creatures
// with flying." (Cloud Sprite) lowers to a RuleEffectCanBlockOnlyCreaturesWith
// rule effect bounded by the flying blocker restriction.
func TestGenerateExecutableCardSourceSelfCanBlockOnlyFlying(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Cloud Watcher",
		Layout:     "normal",
		ManaCost:   "{1}{U}",
		TypeLine:   "Creature — Bird",
		OracleText: "Flying\nThis creature can block only creatures with flying.",
		Colors:     []string{"U"},
		Power:      new("2"),
		Toughness:  new("2"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.RuleEffectCanBlockOnlyCreaturesWith") {
		t.Fatalf("source missing can-block-only-flying rule effect:\n%s", source)
	}
	if !strings.Contains(source, "game.BlockerRestrictionFlying") {
		t.Fatalf("source missing flying blocker restriction:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceRejectsCanBlockOnlyNonFlying confirms the
// recognizer stays narrow and fails closed on a different "can block only"
// restriction it does not support.
func TestGenerateExecutableCardSourceRejectsCanBlockOnlyNonFlying(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Grounded Watcher",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Creature — Spider",
		OracleText: "This creature can block only creatures with power 2 or less.",
		Colors:     []string{"G"},
		Power:      new("2"),
		Toughness:  new("3"),
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "g")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatalf("expected unsupported diagnostics, got generated source:\n%s", source)
	}
	if strings.Contains(source, "game.RuleEffectCanBlockOnlyCreaturesWith") {
		t.Fatalf("unsupported restriction unexpectedly lowered:\n%s", source)
	}
}
