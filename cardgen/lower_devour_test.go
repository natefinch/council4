package cardgen

import (
	"strings"
	"testing"
)

// TestLowerDevourReplacement verifies that the printed Devour keyword lowers to
// a game.DevourReplacement carrying the per-sacrificed-creature multiplier (CR
// 702.81).
func TestLowerDevourReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Thunder-Thrash Elder",
		Layout:     "normal",
		TypeLine:   "Creature — Dinosaur",
		OracleText: "Devour 3 (As this creature enters, you may sacrifice any number of creatures. It enters with three times that many +1/+1 counters on it.)",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if replacement.EntryDevourMultiplier != 3 {
		t.Fatalf("EntryDevourMultiplier = %d, want 3", replacement.EntryDevourMultiplier)
	}
}

// TestGenerateTypedDevourSource verifies that the typed Devour variants lower and
// render to their dedicated constructors carrying the sacrifice filter (CR
// 702.81).
func TestGenerateTypedDevourSource(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		card       *ScryfallCard
		wantSource string
	}{
		{
			name: "artifact",
			card: &ScryfallCard{
				Name:       "Caprichrome",
				Layout:     "normal",
				TypeLine:   "Artifact Creature — Goat",
				OracleText: "Devour artifact 1 (As this creature enters, you may sacrifice any number of artifacts. It enters with that many +1/+1 counters on it.)",
			},
			wantSource: `game.DevourTypeReplacement("As this creature enters, you may sacrifice any number of artifacts, then it enters with 1 +1/+1 counters on it for each artifact sacrificed.", 1, types.Artifact)`,
		},
		{
			name: "land",
			card: &ScryfallCard{
				Name:       "Famished Worldsire",
				Layout:     "normal",
				TypeLine:   "Creature — Elemental",
				OracleText: "Devour land 3 (As this creature enters, you may sacrifice any number of lands. It enters with three times that many +1/+1 counters on it.)",
			},
			wantSource: `game.DevourTypeReplacement("As this creature enters, you may sacrifice any number of lands, then it enters with 3 +1/+1 counters on it for each land sacrificed.", 3, types.Land)`,
		},
		{
			name: "Food",
			card: &ScryfallCard{
				Name:       "Feasting Hobbit",
				Layout:     "normal",
				TypeLine:   "Creature — Halfling",
				OracleText: "Devour Food 3 (As this creature enters, you may sacrifice any number of Foods. It enters with three times that many +1/+1 counters on it.)",
			},
			wantSource: `game.DevourSubtypeReplacement("As this creature enters, you may sacrifice any number of Foods, then it enters with 3 +1/+1 counters on it for each Food sacrificed.", 3, types.Food)`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source, _, err := GenerateExecutableCardSource(test.card, "x")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(source, test.wantSource) {
				t.Fatalf("source missing %q:\n%s", test.wantSource, source)
			}
		})
	}
}
