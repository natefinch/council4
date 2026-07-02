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

// spellCantBlockCard builds a red sorcery bearing oracle as its only text so the
// group-scoped "<group> can't block this turn." spell family can be lowered.
func spellCantBlockCard(name, oracle string) *ScryfallCard {
	return &ScryfallCard{
		Name:       name,
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Sorcery",
		OracleText: oracle,
		Colors:     []string{"R"},
	}
}

// TestGenerateExecutableCardSourceGroupCantBlockKeyword verifies that the
// keyword-filtered group spell "Creatures without flying can't block this turn."
// (Falter, Magmatic Chasm, Seismic Stomp) lowers to a single object-less
// ApplyRule placing an unconditional this-turn RuleEffectCantBlock on every
// non-flying creature, with no controller scope and no protected object.
func TestGenerateExecutableCardSourceGroupCantBlockKeyword(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(
		spellCantBlockCard("Test Falter", "Creatures without flying can't block this turn."), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ApplyRule{",
		"Kind:              game.RuleEffectCantBlock,",
		"PermanentTypes:    []types.Card{types.Creature},",
		"AffectedSelection: game.Selection{ExcludedKeyword: game.Flying},",
		"Duration: game.DurationThisTurn,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "AffectedController") {
		t.Fatalf("all-creatures group must carry no controller scope:\n%s", source)
	}
	if strings.Contains(source, "TargetPermanentReference") || strings.Contains(source, "Object:") {
		t.Fatalf("group can't-block must be object-less:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceGroupCantBlockOpponent verifies that the
// controller-scoped group spell "Creatures your opponents control can't block
// this turn." (Cosmotronic Wave, Hazardous Blast) lowers to an object-less
// ApplyRule scoped by ControllerOpponent to every creature for the turn.
func TestGenerateExecutableCardSourceGroupCantBlockOpponent(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(
		spellCantBlockCard("Test Cosmotronic", "Creatures your opponents control can't block this turn."), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:               game.RuleEffectCantBlock,",
		"AffectedController: game.ControllerOpponent,",
		"PermanentTypes:     []types.Card{types.Creature},",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceGroupCantBlockColor verifies that the
// color-filtered group spell "Green creatures can't block this turn." folds the
// color filter onto the affected Selection.
func TestGenerateExecutableCardSourceGroupCantBlockColor(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(
		spellCantBlockCard("Test Green Falter", "Green creatures can't block this turn."), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "AffectedSelection: game.Selection{ColorsAny: []color.Color{color.Green}},") {
		t.Fatalf("source missing green color filter:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceGroupCantBlockFailsClosed verifies that group
// subjects the rule effect cannot faithfully represent — the monocolored family
// (no runtime color filter) and subtype-filtered groups — fail closed rather
// than widening to every creature.
func TestGenerateExecutableCardSourceGroupCantBlockFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Monocolored creatures can't block this turn.",
		"Cowards can't block this turn.",
		"Creatures with power 2 or less can't block this turn.",
	} {
		_, diagnostics, err := GenerateExecutableCardSource(
			spellCantBlockCard("Test Fail", oracle), "t")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("expected diagnostics for %q, got none", oracle)
		}
	}
}
