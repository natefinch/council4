package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceKickerScaledDamage asserts that a Multikicker
// damage spell that chooses one target plus another target for each time it was
// kicked and deals its amount to each of them (Comet Storm) lowers to a single
// any-target spec carrying CountEqualsKickerPlusOne and one EachTarget Damage
// instruction dealing the spell's X to every chosen target.
func TestGenerateExecutableCardSourceKickerScaledDamage(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Comet Storm",
		Layout:     "normal",
		ManaCost:   "{X}{R}",
		TypeLine:   "Instant",
		OracleText: "Multikicker {1}\nChoose any target, then choose another target for each time this spell was kicked. Comet Storm deals X damage to each of them.",
		Colors:     []string{"R"},
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	wantedSnips := []string{
		"game.KickerKeyword{Cost: cost.Mana{cost.O(1)}, Multi: true}",
		"MinTargets: 1,",
		"MaxTargets: 21,",
		"Constraint: \"any target\",",
		"Allow: game.TargetAllowPermanent | game.TargetAllowPlayer,",
		"CountEqualsKickerPlusOne: true,",
		"Primitive: game.Damage{",
		"Kind: game.DynamicAmountX,",
		"Recipient: game.AnyTargetDamageRecipient(0),",
		"EachTarget: true,",
	}
	// gofmt aligns struct-field colons with runs of spaces; collapse whitespace
	// so the snippet comparison ignores column alignment.
	collapsed := strings.Join(strings.Fields(source), " ")
	for _, wanted := range wantedSnips {
		if !strings.Contains(collapsed, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
