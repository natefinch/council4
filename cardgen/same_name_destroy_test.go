package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableSameNameDestroy exercises the "same name as that <noun>"
// linked-destroy family (issue #2514): a single-target destroy that also removes
// every other battlefield permanent sharing the chosen target's name. The parser
// records the trailing "and all other <type> with the same name as that <noun>"
// clause as the target selector's same-name group, and the lowering emits a
// single Destroy over a SameNamePermanentGroup anchored on the chosen target,
// which includes the target itself.
func TestGenerateExecutableSameNameDestroy(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		typeLine   string
		manaCost   string
		oracle     string
		wantTarget string
		wantGroup  string
	}{
		{
			name:       "Maelstrom Pulse",
			typeLine:   "Sorcery",
			manaCost:   "{1}{B}{G}",
			oracle:     "Destroy target nonland permanent and all other permanents with the same name as that permanent.",
			wantTarget: "Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),",
			wantGroup:  "Group: game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{}),",
		},
		{
			name:       "Wake of Destruction",
			typeLine:   "Sorcery",
			manaCost:   "{3}{R}{R}{R}",
			oracle:     "Destroy target land and all other lands with the same name as that land.",
			wantTarget: "Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),",
			wantGroup:  "Group: game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{RequiredTypes: []types.Card{types.Land}}),",
		},
		{
			name:       "Echoing Ruin",
			typeLine:   "Sorcery",
			manaCost:   "{1}{R}",
			oracle:     "Destroy target artifact and all other artifacts with the same name as that artifact.",
			wantTarget: "Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),",
			wantGroup:  "Group: game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{RequiredTypes: []types.Card{types.Artifact}}),",
		},
		{
			name:       "Echoing Calm",
			typeLine:   "Instant",
			manaCost:   "{1}{W}",
			oracle:     "Destroy target enchantment and all other enchantments with the same name as that enchantment.",
			wantTarget: "Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}),",
			wantGroup:  "Group: game.SameNamePermanentGroup(game.TargetPermanentReference(0), game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				ManaCost:   tc.manaCost,
				TypeLine:   tc.typeLine,
				OracleText: tc.oracle,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "o")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{tc.wantTarget, tc.wantGroup, "Primitive: game.Destroy{"} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}
