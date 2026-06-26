package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutablePluralSubtypeAnthems proves the plural creature-subtype
// anthem group subjects lower end to end: a multi-subtype "you control" anthem
// folds every named subtype into a single controlled-group SubtypesAny modifier,
// and a battlefield-wide tribal lord ("Other <Subtype>s get ...") excludes the
// source from the battlefield-wide subtype group.
func TestGenerateExecutablePluralSubtypeAnthems(t *testing.T) {
	t.Parallel()
	power := "1"
	for name, tc := range map[string]struct {
		card  *ScryfallCard
		wants []string
	}{
		"controlled conjunction": {
			card: &ScryfallCard{
				Name:       "Master Trinketeer",
				Layout:     "normal",
				ManaCost:   "{3}{W}",
				TypeLine:   "Creature — Dwarf Artificer",
				OracleText: "Servos and Thopters you control get +1/+1.",
				Power:      &power,
				Toughness:  &power,
			},
			wants: []string{
				"game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub(\"Servo\"), types.Sub(\"Thopter\")}})",
				"PowerDelta: 1",
				"ToughnessDelta: 1",
			},
		},
		"controlled oxford list": {
			card: &ScryfallCard{
				Name:       "Death-Priest of Myrkul",
				Layout:     "normal",
				ManaCost:   "{1}{B}",
				TypeLine:   "Creature — Tiefling Cleric",
				OracleText: "Skeletons, Vampires, and Zombies you control get +1/+1.",
				Power:      &power,
				Toughness:  &power,
			},
			wants: []string{
				"SubtypesAny: []types.Sub{types.Sub(\"Skeleton\"), types.Sub(\"Vampire\"), types.Sub(\"Zombie\")}",
			},
		},
		"battlefield other single": {
			card: &ScryfallCard{
				Name:       "Goblin King",
				Layout:     "normal",
				ManaCost:   "{1}{R}{R}",
				TypeLine:   "Creature — Goblin",
				OracleText: "Other Goblins get +1/+1 and have mountainwalk.",
				Power:      &power,
				Toughness:  &power,
			},
			wants: []string{
				"game.BattlefieldGroupExcluding(game.Selection{SubtypesAny: []types.Sub{types.Sub(\"Goblin\")}}, game.SourcePermanentReference())",
				"game.LandwalkKeyword{Subtype: types.Mountain}",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(tc.card, "e")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			normalized := strings.Join(strings.Fields(source), " ")
			for _, want := range tc.wants {
				if !strings.Contains(normalized, strings.Join(strings.Fields(want), " ")) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}
