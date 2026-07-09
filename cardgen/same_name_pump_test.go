package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableSameNamePump exercises the "same name as that <noun>"
// power/toughness family: a single-target creature pump that also modifies every
// other creature sharing the chosen target's name until end of turn ("Target
// creature and all other creatures with the same name as that creature get
// -3/-3 until end of turn.", Bile Blight; the Echoing pumps). The parser records
// the trailing "and all other <group> with the same name as that <noun>" clause
// as the target selector's same-name group, and the lowering emits a single
// until-end-of-turn LayerPowerToughnessModify continuous effect over a
// SameNamePermanentGroup anchored on the chosen target, which includes the
// target itself — the group analogue of the single-target ModifyPT pump.
func TestGenerateExecutableSameNamePump(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		manaCost string
		oracle   string
		wantPT   []string
	}{
		{
			name:     "Bile Blight",
			manaCost: "{B}{B}",
			oracle:   "Target creature and all other creatures with the same name as that creature get -3/-3 until end of turn.",
			wantPT:   []string{"PowerDelta:     -3,", "ToughnessDelta: -3,"},
		},
		{
			name:     "Echoing Courage",
			manaCost: "{1}{G}",
			oracle:   "Target creature and all other creatures with the same name as that creature get +2/+2 until end of turn.",
			wantPT:   []string{"PowerDelta:     2,", "ToughnessDelta: 2,"},
		},
		{
			name:     "Echoing Decay",
			manaCost: "{1}{B}",
			oracle:   "Target creature and all other creatures with the same name as that creature get -2/-2 until end of turn.",
			wantPT:   []string{"PowerDelta:     -2,", "ToughnessDelta: -2,"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				ManaCost:   tc.manaCost,
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "o")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			wanted := []string{
				"Constraint: \"target creature and all other creatures with the same name as that creature\",",
				"Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),",
				"Primitive: game.ApplyContinuous{",
				"Layer:          game.LayerPowerToughnessModify,",
				"Group:          game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{RequiredTypes: []types.Card{types.Creature}}),",
				"Duration: game.DurationUntilEndOfTurn,",
			}
			wanted = append(wanted, tc.wantPT...)
			for _, want := range wanted {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}
