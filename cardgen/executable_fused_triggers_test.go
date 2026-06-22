package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceFusedTrigger verifies that a fused
// "When ~ enters and whenever <event>, <effect>" ability lowers to two
// independent triggered abilities sharing the effect: one firing when the source
// enters, one firing on the joined event.
func TestGenerateExecutableCardSourceFusedTrigger(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Beanstalk Sample",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "When this enchantment enters and whenever you cast a spell with mana value 5 or greater, draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if got := strings.Count(source, "Trigger: game.TriggerCondition"); got != 2 {
		t.Fatalf("want two triggered abilities, got %d:\n%s", got, source)
	}
	for _, wanted := range []string{
		"game.EventPermanentEnteredBattlefield",
		"game.TriggerSourceSelf",
		"game.EventSpellCast",
		"Primitive: game.Draw",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if got := strings.Count(source, "Primitive: game.Draw"); got != 2 {
		t.Fatalf("want both abilities to draw, got %d draw effects:\n%s", got, source)
	}
}
