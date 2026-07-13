package cardgen

import (
	"strings"
	"testing"
)

const explosiveDerailmentSpreeText = "Spree (Choose one or more additional costs.)\n" +
	"+ {2} — Explosive Derailment deals 4 damage to target creature.\n" +
	"+ {2} — Destroy target artifact."

func TestGenerateSpreeSpellSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Explosive Derailment",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
		OracleText: explosiveDerailmentSpreeText,
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Modes: []game.Mode{",
		"Cost: opt.Val(cost.Mana{cost.O(2)}),",
		"MinModes: 1,",
		"MaxModes: 2,",
		"Primitive: game.Damage{",
		"Primitive: game.Destroy",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

func TestGenerateSpreeRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()
	// The second Spree option detains a creature, an effect the executable backend
	// does not lower, so the whole card must fail closed even though the Spree
	// structure and the first option are recognized.
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Explosive Derailment",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{R}",
		OracleText: "Spree (Choose one or more additional costs.)\n" +
			"+ {2} — Explosive Derailment deals 4 damage to target creature.\n" +
			"+ {2} — Detain target creature.",
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("diagnostics = none; want the unsupported detain mode to fail closed")
	}
}

// TestGenerateReturnTheFavorSource proves Return the Favor (EDHREC #893, #3019)
// now lowers end to end. Its first Spree option copies a target drawn from a
// mixed spell/ability stack-object union whose spell arm is restricted to
// instant and sorcery ("instant spell, sorcery spell, activated ability, or
// triggered ability") and may choose new targets for the copy; its second
// option retargets a spell or ability with a single target. The union is modeled
// generically through StackObjectKinds plus SpellCardTypesAny, not a
// card-specific target.
func TestGenerateReturnTheFavorSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Return the Favor",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{R}{R}",
		OracleText: "Spree (Choose one or more additional costs.)\n" +
			"+ {1} — Copy target instant spell, sorcery spell, activated ability, or triggered ability. You may choose new targets for the copy.\n" +
			"+ {1} — Change the target of target spell or ability with a single target.",
	}, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Modes: []game.Mode{",
		"Cost: opt.Val(cost.Mana{cost.O(1)}),",
		"SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},",
		"StackObjectKinds:  []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},",
		"Primitive: game.CopyStackObject{",
		"MayChooseNewTargets: true,",
		"Primitive: game.ChooseNewTargets{",
		"MinModes: 1,",
		"MaxModes: 2,",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateSpreeOptionMissingCostFailsClosed proves the compiler/lowering
// require every Spree option to carry its own additional cost (CR 702.171). A
// Spree option printed without a "+ {cost} —" clause is structurally malformed
// and must fail closed rather than lower a costless option.
func TestGenerateSpreeOptionMissingCostFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Malformed Spree",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{R}",
		OracleText: "Spree (Choose one or more additional costs.)\n" +
			"+ {2} — Malformed Spree deals 4 damage to target creature.\n" +
			"+ Destroy target artifact.",
	}, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("diagnostics = none; want the costless Spree option to fail closed")
	}
}
