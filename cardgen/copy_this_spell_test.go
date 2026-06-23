package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateCopyThisSpellSource covers the "Copy this spell." resolving
// self-copy primitive: the bare clause, the "you may choose new targets for the
// copy" rider, and Sevinne's Reclamation, whose copy is gated on the spell
// having been cast from a graveyard and carries the conjoined singular
// new-target rider.
func TestGenerateCopyThisSpellSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		typeLine string
		manaCost string
		oracle   string
		want     []string
	}{
		{
			name:     "Copythis Bare",
			typeLine: "Instant",
			manaCost: "{R}",
			oracle:   "Copy this spell.",
			want: []string{
				"Primitive: game.CopyStackObject{",
				"Object:              game.ResolvingStackObjectReference(),",
			},
		},
		{
			name:     "Copythis New Targets",
			typeLine: "Instant",
			manaCost: "{U}",
			oracle:   "Copy this spell. You may choose new targets for the copy.",
			want: []string{
				"Object:              game.ResolvingStackObjectReference(),",
				"MayChooseNewTargets: true,",
			},
		},
		{
			name:     "Sevinne's Reclamation",
			typeLine: "Sorcery",
			manaCost: "{2}{W}",
			oracle: "Return target permanent card with mana value 3 or less from your graveyard to the battlefield. " +
				"If this spell was cast from a graveyard, you may copy this spell and may choose a new target for the copy.\n" +
				"Flashback {4}{W} (You may cast this card from your graveyard for its flashback cost. Then exile it.)",
			want: []string{
				"Primitive: game.CopyStackObject{",
				"Object:              game.ResolvingStackObjectReference(),",
				"MayChooseNewTargets: true,",
				"CastFromZone: opt.Val(zone.Graveyard),",
				"Optional: true,",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				ManaCost:   test.manaCost,
				OracleText: test.oracle,
			}, "c")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.want {
				if !strings.Contains(spaceCollapsed(source), spaceCollapsed(want)) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}
