package cardgen

import (
	"strings"
	"testing"
)

// TestInheritedRemovalTargetRefSubtypeLead proves the inherited-removal-target
// recipient helpers accept bare subtype ("Destroy target Mountain.") and compound
// type ("Destroy target Plains or Island.") leads, not just the plain card-type
// selectors. Each clause resolves the recipient to the destroyed permanent's
// controller through a permanent reference, the same shape a plain "target land"
// lead produces.
func TestInheritedRemovalTargetRefSubtypeLead(t *testing.T) {
	t.Parallel()
	const recipient = "game.ObjectControllerReference(game.TargetPermanentReference(0))"
	for _, tc := range []struct {
		name      string
		oracle    string
		primitive string
	}{
		// Subtype lead, damage rider ("that land's controller").
		{
			"Test Peak Burn",
			"Destroy target Mountain. Test Peak Burn deals 3 damage to that land's controller.",
			"game.Damage{",
		},
		// Compound type lead, damage rider.
		{
			"Test Dual Burn",
			"Destroy target Plains or Island. Test Dual Burn deals 3 damage to that land's controller.",
			"game.Damage{",
		},
		// Subtype lead, "Its controller loses N life" rider.
		{
			"Test Peak Drain",
			"Destroy target Mountain. Its controller loses 2 life.",
			"game.LoseLife{",
		},
	} {
		source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       tc.name,
			Layout:     "normal",
			ManaCost:   "{2}{R}",
			TypeLine:   "Sorcery",
			OracleText: tc.oracle,
		}, "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("%q: diagnostics = %#v", tc.oracle, diagnostics)
		}
		for _, want := range []string{
			"Primitive: " + tc.primitive,
			recipient,
		} {
			if !strings.Contains(source, want) {
				t.Fatalf("%q: source missing %q:\n%s", tc.oracle, want, source)
			}
		}
	}
}
