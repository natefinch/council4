package cardgen

import (
	"strings"
	"testing"
)

const calamityGallopingInfernoOracle = "Haste\n" +
	"Whenever Calamity attacks while saddled, choose a nonlegendary creature that saddled it this turn and create a tapped and attacking token that's a copy of it. Sacrifice that token at the beginning of the next end step. Repeat this process once.\n" +
	"Saddle 1"

func TestGenerateCalamityComposableMechanics(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Calamity, Galloping Inferno",
		Layout:     "normal",
		ManaCost:   "{4}{R}{R}",
		TypeLine:   "Legendary Creature — Horse Mount",
		OracleText: calamityGallopingInfernoOracle,
		Colors:     []string{"R"},
		Power:      new("4"),
		Toughness:  new("6"),
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.RepeatProcess{",
		"Times: game.Fixed(2)",
		"game.SaddleContributorsGroup(",
		"ExcludedSupertype: types.Legendary",
		"Source: game.TokenCopySourceChosenFromGroup",
		"EntryTapped:        true",
		"AttackSameAsSource: true",
		"PublishLinked:",
		"CapturedObjectGroup:",
		"Primitive: game.Sacrifice{",
		"Group: game.CapturedObjectsGroup()",
	} {
		if !strings.Contains(source, wanted) {
			t.Errorf("generated source missing %q:\n%s", wanted, source)
		}
	}
}
