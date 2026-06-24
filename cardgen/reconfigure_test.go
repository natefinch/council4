package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

const reconfigureReminder = "({2}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)"

// TestLowerReconfigureAbility verifies a printed "Reconfigure {N}" keyword on an
// Equipment creature lowers to a single ReconfigureActivatedAbility carrying the
// Reconfigure keyword identity and mana cost.
func TestLowerReconfigureAbility(t *testing.T) {
	card := &ScryfallCard{
		Name:       "Lizard Blades",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Equipment Lizard",
		ManaCost:   "{1}{R}",
		OracleText: "Double strike\nEquipped creature has double strike.\nReconfigure {2} " + reconfigureReminder,
	}
	face := lowerSingleFace(t, card)
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	manaCost, ok := game.ActivatedBodyReconfigureCost(&face.ActivatedAbilities[0])
	if !ok {
		t.Fatalf("activated ability is not a Reconfigure ability: %+v", face.ActivatedAbilities[0])
	}
	if got := manaCost.String(); got != "{2}" {
		t.Fatalf("reconfigure cost = %q, want {2}", got)
	}
}

// TestGenerateExecutableCardSourceReconfigure verifies the Reconfigure keyword
// round-trips to a ReconfigureActivatedAbility call in generated source with no
// diagnostics.
func TestGenerateExecutableCardSourceReconfigure(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Rabbit Battery",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Equipment Rabbit",
		ManaCost:   "{R}",
		OracleText: "Haste\nEquipped creature gets +1/+1 and has haste.\nReconfigure {R} ({R}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "r")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"ActivatedAbilities: []game.ActivatedAbility",
		"game.ReconfigureActivatedAbility(cost.Mana{cost.R})",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceReconfigureEmDashUnsupported verifies the
// non-mana em-dash Reconfigure form ("Reconfigure—Pay {2} or {E}{E}{E}.")
// fails closed rather than lowering to an attach ability.
func TestGenerateExecutableCardSourceReconfigureEmDashUnsupported(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Razorfield Ripper",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Equipment Cat",
		ManaCost:   "{4}",
		OracleText: "Reconfigure—Pay {2} or {E}{E}{E}. (Pay {2} or {E}{E}{E}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)",
	}
	_, diagnostics, _ := GenerateExecutableCardSource(card, "r")
	if len(diagnostics) == 0 {
		t.Fatal("expected diagnostics for unsupported em-dash Reconfigure, got none")
	}
}
