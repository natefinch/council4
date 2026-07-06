package cardgen

import (
	"strings"
	"testing"
)

const collectiveResistanceEscalateText = "Escalate {G} (Pay this cost for each mode chosen beyond the first.)\n" +
	"Choose one or more —\n" +
	"• Destroy target artifact.\n" +
	"• Destroy target enchantment.\n" +
	"• Target creature gains hexproof and indestructible until end of turn."

func TestGenerateEscalateSpellSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Collective Resistance",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{G}",
		OracleText: collectiveResistanceEscalateText,
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"Modes: []game.Mode{",
		"EscalateCost: opt.Val(cost.Mana{cost.G}),",
		"MinModes:",
		"MaxModes:",
		"Primitive: game.Destroy",
		"AddKeywords: []game.Keyword{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
	for _, want := range []string{"MinModes: 1", "MaxModes: 3"} {
		if !strings.Contains(spaceCollapsed(source), want) {
			t.Fatalf("collapsed source missing %q", want)
		}
	}
}

// spaceCollapsed collapses runs of spaces so assertions about rendered struct
// fields do not depend on gofmt's column alignment.
func spaceCollapsed(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func TestGenerateEscalateRejectsUnsupportedMode(t *testing.T) {
	t.Parallel()
	// The third option detains a creature, an effect the executable backend does
	// not lower, so the whole Escalate card must fail closed even though the
	// keyword header and the first two options are recognized.
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Collective Resistance",
		Layout:   "normal",
		TypeLine: "Instant",
		ManaCost: "{1}{G}",
		OracleText: "Escalate {G} (Pay this cost for each mode chosen beyond the first.)\n" +
			"Choose one or more —\n" +
			"• Destroy target artifact.\n" +
			"• Detain target creature.",
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("diagnostics = none; want the unsupported detain mode to fail closed")
	}
}
