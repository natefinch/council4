package cardgen

import (
	"strings"
	"testing"
)

const nacatlWarPrideOracle = "This creature must be blocked by exactly one creature if able.\n" +
	"Whenever this creature attacks, create X tokens that are copies of it and that are tapped and attacking, where X is the number of creatures defending player controls. Exile the tokens at the beginning of the next end step."

func TestGenerateNacatlWarPrideExecutableSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Nacatl War-Pride",
		Layout:     "normal",
		ManaCost:   "{3}{G}{G}{G}",
		TypeLine:   "Creature — Cat Warrior",
		OracleText: nacatlWarPrideOracle,
		Colors:     []string{"G"},
		Power:      new("3"),
		Toughness:  new("3"),
	}, "n")
	if err != nil {
		t.Fatal(err)
	}

	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.MustBeBlockedByExactlyOneStaticBody",
		"game.PlayerControlledGroup(game.DefendingPlayerReference()",
		"Object: game.SourcePermanentReference()",
		"EntryAttackingDefender: opt.Val(game.DefendingPlayerReference())",
		"PublishLinked:",
		"CapturedObjectGroup:",
		"Group: game.CapturedObjectsGroup()",
	} {
		if !strings.Contains(source, wanted) {
			t.Errorf("generated source missing %q:\n%s", wanted, source)
		}
	}
}

func TestDefendingPlayerCountGroupPumpFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Mercadia's Downfall",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Instant",
		OracleText: "Each attacking creature gets +1/+0 until end of turn for each nonbasic land defending player controls.",
		Colors:     []string{"R"},
	}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("group pump with per-attacker defending-player count unexpectedly lowered")
	}
}
