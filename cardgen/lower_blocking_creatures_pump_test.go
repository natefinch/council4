package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// assertBlockingCreaturesPump checks a lowered ModifyPT scales both sides by the
// blocking-creatures count bound to the pumped permanent.
func assertBlockingCreaturesPump(t *testing.T, modify game.ModifyPT, multiplier int) {
	t.Helper()
	if modify.Object != game.EventPermanentReference() {
		t.Fatalf("object = %+v, want event permanent reference", modify.Object)
	}
	for _, side := range []struct {
		name     string
		quantity game.Quantity
	}{{"power", modify.PowerDelta}, {"toughness", modify.ToughnessDelta}} {
		dynamic := side.quantity.DynamicAmount()
		if !dynamic.Exists ||
			dynamic.Val.Kind != game.DynamicAmountBlockingCreatures ||
			dynamic.Val.Multiplier != multiplier ||
			dynamic.Val.Object != game.EventPermanentReference() {
			t.Fatalf("%s delta = %+v, want blocking-creatures multiplier %d bound to the pumped permanent",
				side.name, side.quantity, multiplier)
		}
	}
}

// TestLowerBlockingCreaturesSelfPump lowers the self-form "Whenever this creature
// becomes blocked, it gets +2/+2 … for each creature blocking it." (Rabid
// Elephant, Gang of Elk). The pump addresses the just-blocked permanent and
// scales by the count of creatures blocking it.
func TestLowerBlockingCreaturesSelfPump(t *testing.T) {
	t.Parallel()
	modify := referencedDynamicModifyPT(t,
		"Creature — Elephant",
		"Whenever this creature becomes blocked, it gets +2/+2 until end of turn for each creature blocking it.",
		false)
	assertBlockingCreaturesPump(t, modify, 2)
}

// TestLowerBlockingCreaturesOtherCreaturePump lowers the other-creature form
// "Whenever a Beast becomes blocked, it gets +1/+1 … for each creature blocking
// it." (Berserk Murlodont). The "it" names the triggering blocked creature, so
// the count and the pump both bind to that permanent.
func TestLowerBlockingCreaturesOtherCreaturePump(t *testing.T) {
	t.Parallel()
	modify := referencedDynamicModifyPT(t,
		"Creature — Beast",
		"Whenever a Beast becomes blocked, it gets +1/+1 until end of turn for each creature blocking it.",
		false)
	assertBlockingCreaturesPump(t, modify, 1)
}

// TestLowerBlockingCreaturesKeywordRiderRejected keeps a blocking-creatures pump
// that also grants a keyword fail closed: the pump lowering models only the bare
// power/toughness change, so a "… and gains …" rider must not be silently
// dropped.
func TestLowerBlockingCreaturesKeywordRiderRejected(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Blocking Pump Rider",
		Layout:     "normal",
		TypeLine:   "Creature — Elephant",
		OracleText: "Whenever this creature becomes blocked, it gets +2/+2 and gains trample until end of turn for each creature blocking it.",
		Games:      []string{"paper"},
		Legalities: map[string]string{"legacy": "legal"},
	})
	if len(diagnostics) == 0 {
		t.Fatalf("diagnostics = %+v, want the keyword rider to fail closed", diagnostics)
	}
}
