package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

const displacedDinosaursOracle = "As a historic permanent you control enters, it becomes a 7/7 Dinosaur creature in addition to its other types."

// TestLowerGroupEntersBecomesReplacement proves Displaced Dinosaurs lowers to a
// continuous EntersBecomesGroupReplacement carrying the historic you-control
// filter, the added Creature type and Dinosaur subtype, and the 7/7 base P/T.
func TestLowerGroupEntersBecomesReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Displaced Dinosaurs",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: displacedDinosaursOracle,
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if !replacement.EntersBecomesCharacteristic {
		t.Fatalf("replacement is not an enters-becomes characteristic replacement: %#v", replacement)
	}
	if replacement.ControllerFilter != game.TriggerControllerYou {
		t.Errorf("controller filter = %v, want you", replacement.ControllerFilter)
	}
	if replacement.EntersBecomesSelection == nil || len(replacement.EntersBecomesSelection.AnyOf) == 0 {
		t.Errorf("entrant selection = %#v, want a historic AnyOf filter", replacement.EntersBecomesSelection)
	}
	if !slices.Equal(replacement.EntersBecomesAddTypes, []types.Card{types.Creature}) {
		t.Errorf("add types = %v, want [Creature]", replacement.EntersBecomesAddTypes)
	}
	if !slices.Equal(replacement.EntersBecomesAddSubtypes, []types.Sub{types.Dinosaur}) {
		t.Errorf("add subtypes = %v, want [Dinosaur]", replacement.EntersBecomesAddSubtypes)
	}
	if !replacement.EntersBecomesBasePower.Exists || replacement.EntersBecomesBasePower.Val != 7 {
		t.Errorf("base power = %v, want 7", replacement.EntersBecomesBasePower)
	}
	if !replacement.EntersBecomesBaseToughness.Exists || replacement.EntersBecomesBaseToughness.Val != 7 {
		t.Errorf("base toughness = %v, want 7", replacement.EntersBecomesBaseToughness)
	}
}

// TestRenderGroupEntersBecomesReplacement proves the lowered replacement renders
// back to a compilable EntersBecomesGroupReplacement constructor call with the
// full EntersBecomesGroupParams literal.
func TestRenderGroupEntersBecomesReplacement(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Displaced Dinosaurs",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: displacedDinosaursOracle,
	})
	ability := face.ReplacementAbilities[0]
	rendered, err := (Renderer{}).renderReplacementAbility(newRenderCtx(), &ability)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"game.EntersBecomesGroupReplacement(",
		"game.EntersBecomesGroupParams{",
		"Controller: game.TriggerControllerYou",
		"Historic: true",
		"AddTypes: []types.Card{types.Creature}",
		"AddSubtypes: []types.Sub{types.Dinosaur}",
		"BasePower: opt.Val(7)",
		"BaseToughness: opt.Val(7)",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("rendered replacement missing %q:\n%s", want, rendered)
		}
	}
}

// TestGenerateExecutableCardSourceDisplacedDinosaurs proves the full pipeline
// generates diagnostic-free executable source for Displaced Dinosaurs.
func TestGenerateExecutableCardSourceDisplacedDinosaurs(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Displaced Dinosaurs",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: displacedDinosaursOracle,
	}, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if !strings.Contains(source, "game.EntersBecomesGroupReplacement(") {
		t.Fatalf("generated source missing enters-becomes replacement:\n%s", source)
	}
}
