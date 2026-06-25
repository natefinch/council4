package cardgen

import (
	"strings"
	"testing"
)

func cantBlockTestCard(oracle string) *ScryfallCard {
	power := "2"
	toughness := "2"
	return &ScryfallCard{
		Name:       "Test Cant Block",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Creature — Beast",
		OracleText: oracle,
		Colors:     []string{"G"},
		Power:      &power,
		Toughness:  &toughness,
	}
}

// TestGenerateExecutableCardSourceCreaturesCantBlockSource verifies that
// "Creatures with power less than this creature's power can't block it."
// (Wandering Wolf, Aura Gnarlid) lowers to a battlefield can't-block rule effect
// whose restricted blockers are filtered by the source's power and whose
// protected object is the source itself.
func TestGenerateExecutableCardSourceCreaturesCantBlockSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(
		cantBlockTestCard("Creatures with power less than this creature's power can't block it."), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:              game.RuleEffectCantBlock,",
		"PermanentTypes:    []types.Card{types.Creature},",
		"AffectedSelection: game.Selection{PowerLessThanSource: true},",
		"BlockedSource:     true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCreaturesCantBlockControlled verifies that
// "Creatures with power less than this creature's power can't block creatures you
// control." (Champion of Lambholt) lowers to a battlefield can't-block rule
// effect whose protected object is the controller's creatures.
func TestGenerateExecutableCardSourceCreaturesCantBlockControlled(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(
		cantBlockTestCard("Creatures with power less than this creature's power can't block creatures you control."), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:              game.RuleEffectCantBlock,",
		"AffectedSelection: game.Selection{PowerLessThanSource: true},",
		"BlockedSelection:  game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceCreaturesCantBlockUnconditional verifies that
// the unconditional "Creatures can't block." (Bedlam) lowers to a battlefield
// can't-block rule effect with no blocker filter and no protected-object scope.
func TestGenerateExecutableCardSourceCreaturesCantBlockUnconditional(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(
		cantBlockTestCard("Creatures can't block."), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:           game.RuleEffectCantBlock,",
		"PermanentTypes: []types.Card{types.Creature},",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "BlockedSource") || strings.Contains(source, "BlockedSelection") {
		t.Fatalf("unconditional can't-block source should carry no protected object:\n%s", source)
	}
}
