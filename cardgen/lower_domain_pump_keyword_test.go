package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// applyContinuousFromSequence returns the lone ApplyContinuous primitive in a
// single-mode ability content's first sequence slot.
func applyContinuousFromSequence(t *testing.T, content game.AbilityContent) game.ApplyContinuous {
	t.Helper()
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
		t.Fatalf("content = %+v, want one mode with one instruction", content)
	}
	apply, ok := content.Modes[0].Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", content.Modes[0].Sequence[0].Primitive)
	}
	return apply
}

// TestLowerDomainPumpKeywordSpell verifies the combined "Target creature gets
// +X/+X and gains <keyword> until end of turn, where X is the number of basic
// land types among lands you control." (domain) pump lowers to one
// ApplyContinuous carrying a dynamic power/toughness layer and a keyword layer.
// The trailing "where X is …" clause binds to the gain effect, so the lowering
// must resolve X from there for the pump.
func TestLowerDomainPumpKeywordSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Domain Buff",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{G}",
		OracleText: "Target creature you control gets +X/+X and gains trample until end of turn, where X is the number of basic land types among lands you control.",
	})
	apply := applyContinuousFromSequence(t, face.SpellAbility.Val)
	if len(apply.ContinuousEffects) != 2 {
		t.Fatalf("continuous effects = %d, want 2", len(apply.ContinuousEffects))
	}
	pt := apply.ContinuousEffects[0]
	if pt.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("layer[0] = %v, want power/toughness modify", pt.Layer)
	}
	for _, side := range []struct {
		name    string
		dynamic game.Quantity
	}{
		{"power", game.Dynamic(pt.PowerDeltaDynamic.Val)},
		{"toughness", game.Dynamic(pt.ToughnessDeltaDynamic.Val)},
	} {
		dyn := side.dynamic.DynamicAmount()
		if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountControllerBasicLandTypeCount {
			t.Fatalf("%s delta dynamic = %+v, want controller basic land type count", side.name, side.dynamic)
		}
	}
	keywordLayer := apply.ContinuousEffects[1]
	if keywordLayer.Layer != game.LayerAbility ||
		len(keywordLayer.AddKeywords) != 1 ||
		keywordLayer.AddKeywords[0] != game.Trample {
		t.Fatalf("keyword layer = %+v, want add trample", keywordLayer)
	}
	if apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want until end of turn", apply.Duration)
	}
}

// TestLowerFixedPumpKeywordStillSupported guards against a regression in the
// pre-existing fixed-delta combined pump that the dynamic generalization must
// preserve byte-for-byte.
func TestLowerFixedPumpKeywordStillSupported(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Fixed Buff",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{1}{G}",
		OracleText: "Target creature gets +1/+1 and gains trample until end of turn.",
	})
	apply := applyContinuousFromSequence(t, face.SpellAbility.Val)
	pt := apply.ContinuousEffects[0]
	if pt.PowerDelta != 1 || pt.ToughnessDelta != 1 ||
		pt.PowerDeltaDynamic.Exists || pt.ToughnessDeltaDynamic.Exists {
		t.Fatalf("power/toughness layer = %+v, want fixed +1/+1", pt)
	}
}

// TestLowerWeatherseedTreatySaga compiles The Weatherseed Treaty end to end,
// confirming the read-ahead Saga lowers all three chapters once the domain
// pump+keyword chapter III is supported (the read-ahead sacrifice-chapter check
// depends on every chapter lowering).
func TestLowerWeatherseedTreatySaga(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "The Weatherseed Treaty",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		ManaCost: "{2}{G}",
		OracleText: "Read ahead (Choose a chapter and start with that many lore counters. Add one after your draw step. Skipped chapters don't trigger. Sacrifice after III.)\n" +
			"I — Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.\n" +
			"II — Create a 1/1 green Saproling creature token.\n" +
			"III — Domain — Target creature you control gets +X/+X and gains trample until end of turn, where X is the number of basic land types among lands you control.",
	})
	if len(face.ChapterAbilities) != 3 {
		t.Fatalf("chapter abilities = %d, want 3", len(face.ChapterAbilities))
	}
}
