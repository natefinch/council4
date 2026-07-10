package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerMyojinCastFromHandDivinityCounter verifies the original Myojin
// wording lowers to a conditional enters-with-counters replacement gated on the
// entering permanent having been cast by its controller from their hand.
func TestLowerMyojinCastFromHandDivinityCounter(t *testing.T) {
	t.Parallel()
	const oracleText = "Myojin of Seeing Winds enters with a divinity counter on it if you cast it from your hand."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Myojin of Seeing Winds",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Spirit",
		ManaCost:   "{7}{U}{U}{U}",
		OracleText: oracleText,
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("replacement abilities = %d, want 1", len(face.ReplacementAbilities))
	}

	replacement := face.ReplacementAbilities[0].Replacement
	if len(replacement.EntersWithCounters) != 1 {
		t.Fatalf("counter placements = %d, want 1", len(replacement.EntersWithCounters))
	}
	placement := replacement.EntersWithCounters[0]
	if placement.Kind != counter.Divinity || placement.Amount != 1 {
		t.Fatalf("counter placement = %#v, want one divinity counter", placement)
	}
	if !replacement.Condition.Exists ||
		!replacement.Condition.Val.EventPermanentWasCastFromControllerHand {
		t.Fatalf("condition = %#v, want cast-from-controller-hand gate", replacement.Condition)
	}
}

// TestLowerCastFromHandCountersWithKeywordFailsClosed verifies the unsupported
// combined "counters and with flying" form does not silently drop the keyword.
func TestLowerCastFromHandCountersWithKeywordFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Combined Entry",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		ManaCost:   "{3}{U}",
		OracleText: "Test Combined Entry enters with two +1/+1 counters on it and with flying if you cast it from your hand.",
	})
}
