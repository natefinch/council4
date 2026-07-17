package cardgen

import (
	"strings"
	"testing"
)

const flamerushRiderOracle = "Whenever this creature attacks, create a token that's a copy of another target attacking creature and that's tapped and attacking. Exile the token at end of combat.\n" +
	"Dash {2}{R}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)"

func TestGenerateFlamerushRiderComposableMechanics(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Flamerush Rider",
		Layout:     "normal",
		ManaCost:   "{4}{R}",
		TypeLine:   "Creature — Human Warrior",
		OracleText: flamerushRiderOracle,
		Colors:     []string{"R"},
		Power:      new("3"),
		Toughness:  new("3"),
	}, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"CombatState: game.CombatStateAttacking",
		"ExcludeSource: true",
		"Source: game.TokenCopySourceObject",
		"Object: game.TargetPermanentReference(0)",
		"EntryTapped:        true",
		"AttackSameAsObject: opt.Val(game.TargetPermanentReference(0))",
		"PublishLinked:",
		"Timing:              game.DelayedAtEndOfCombat",
		"CapturedObjectGroup:",
		"Group: game.CapturedObjectsGroup()",
		"Mechanic: cost.AlternativeMechanicDash",
		"game.DashTriggeredAbility()",
	} {
		if !strings.Contains(source, wanted) {
			t.Errorf("generated source missing %q:\n%s", wanted, source)
		}
	}
}
