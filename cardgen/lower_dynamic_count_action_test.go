package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerWhereXInvestigateSpell proves a standalone investigate whose count is
// a "where X is <count>" clause lowers to game.Investigate with a dynamic
// count-selector quantity, the form Ethereal Investigator uses ("investigate X
// times, where X is the number of opponents you have"). It previously failed
// closed because investigate accepted only a fixed literal amount.
func TestLowerWhereXInvestigateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Investigate X",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Investigate X times, where X is the number of artifacts you control.",
	})
	investigate, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Investigate)
	if !ok {
		t.Fatalf("primitive = %T, want game.Investigate", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	dyn := investigate.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountCountSelector {
		t.Fatalf("investigate.Amount = %+v, want dynamic count selector", investigate.Amount)
	}
}

// TestLowerForEachInvestigateSpell proves a standalone "investigate for each
// <count>" clause lowers to a dynamic count-selector investigate, the multiplier
// form Serene Sleuth and Teysa use without a where-X variable.
func TestLowerForEachInvestigateSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Investigate Each",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Investigate for each creature you control.",
	})
	if _, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Investigate); !ok {
		t.Fatalf("primitive = %T, want game.Investigate", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
}

// TestLowerForEachScrySpell proves a fixed-multiplier "Scry N for each <count>"
// clause lowers to a dynamic controller scry, broadening the controller
// scry/surveil family beyond its where-X variant.
func TestLowerForEachScrySpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Scry Each",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Scry 1 for each opponent.",
	})
	scry, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Scry)
	if !ok {
		t.Fatalf("primitive = %T, want game.Scry", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if scry.Player != game.ControllerReference() || !scry.Amount.DynamicAmount().Exists {
		t.Fatalf("scry = %+v, want controller dynamic", scry)
	}
}
