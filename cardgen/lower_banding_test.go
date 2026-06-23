package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerIndefiniteKeywordGrantSpell covers Part 1: a spell that grants a
// keyword to a target creature with no stated duration lowers to an anchored
// ApplyContinuous carrying DurationPermanent (the never-expiring grant).
func TestLowerIndefiniteKeywordGrantSpell(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		oracle  string
		keyword game.Keyword
	}{
		{"first strike", "Target creature gains first strike.", game.FirstStrike},
		{"trample", "Target creature gains trample.", game.Trample},
		{"banding", "Target creature gains banding.", game.Banding},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Indefinite Grant",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			})
			mode := face.SpellAbility.Val.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(mode.Targets))
			}
			apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
			if !ok {
				t.Fatalf("primitive = %T, want game.ApplyContinuous", mode.Sequence[0].Primitive)
			}
			if apply.Duration != game.DurationPermanent {
				t.Fatalf("duration = %v, want DurationPermanent", apply.Duration)
			}
			effect := apply.ContinuousEffects[0]
			if effect.Layer != game.LayerAbility ||
				!reflect.DeepEqual(effect.AddKeywords, []game.Keyword{tc.keyword}) {
				t.Fatalf("continuous effect = %+v, want %v keyword grant", effect, tc.keyword)
			}
		})
	}
}

// TestLowerKeywordChoiceGrantSpell covers Part 2 and Part 3: a spell that grants
// one of several listed keywords (including banding) to a target creature lowers
// to a flat modal AbilityContent with one mode per keyword choice, a single
// shared target, and an exactly-one mode selection.
func TestLowerKeywordChoiceGrantSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Keyword Choice",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gains banding, first strike, or trample.",
	})
	content := face.SpellAbility.Val
	if len(content.SharedTargets) != 1 {
		t.Fatalf("shared targets = %d, want 1", len(content.SharedTargets))
	}
	if content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("modes range = [%d,%d], want exactly one", content.MinModes, content.MaxModes)
	}
	wantKeywords := []game.Keyword{game.Banding, game.FirstStrike, game.Trample}
	if len(content.Modes) != len(wantKeywords) {
		t.Fatalf("modes = %d, want %d", len(content.Modes), len(wantKeywords))
	}
	for i, want := range wantKeywords {
		mode := content.Modes[i]
		apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
		if !ok {
			t.Fatalf("mode %d primitive = %T, want game.ApplyContinuous", i, mode.Sequence[0].Primitive)
		}
		if apply.Duration != game.DurationPermanent {
			t.Fatalf("mode %d duration = %v, want DurationPermanent", i, apply.Duration)
		}
		if !reflect.DeepEqual(apply.ContinuousEffects[0].AddKeywords, []game.Keyword{want}) {
			t.Fatalf("mode %d keywords = %v, want %v", i, apply.ContinuousEffects[0].AddKeywords, want)
		}
	}
}

// TestLowerNaturesBlessing covers Part 4: the goal card lowers to a single
// modal activated ability that shares one creature target across a +1/+1
// counter mode and three indefinite keyword-grant modes (banding, first strike,
// trample), with an exactly-one mode selection.
func TestLowerNaturesBlessing(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Nature's Blessing",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{2}{G}{W}",
		OracleText: "{G}{W}, Discard a card: Put a +1/+1 counter on target creature or that creature gains banding, first strike, or trample. (This effect lasts indefinitely.)",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	content := face.ActivatedAbilities[0].Content
	if len(content.SharedTargets) != 1 {
		t.Fatalf("shared targets = %d, want 1", len(content.SharedTargets))
	}
	if content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("modes range = [%d,%d], want exactly one", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != 4 {
		t.Fatalf("modes = %d, want 4 (counter + three keyword grants)", len(content.Modes))
	}
	counterMode := content.Modes[0]
	if _, ok := counterMode.Sequence[0].Primitive.(game.AddCounter); !ok {
		t.Fatalf("mode 0 primitive = %T, want game.AddCounter", counterMode.Sequence[0].Primitive)
	}
	wantKeywords := []game.Keyword{game.Banding, game.FirstStrike, game.Trample}
	for i, want := range wantKeywords {
		mode := content.Modes[i+1]
		apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
		if !ok {
			t.Fatalf("mode %d primitive = %T, want game.ApplyContinuous", i+1, mode.Sequence[0].Primitive)
		}
		if apply.Duration != game.DurationPermanent {
			t.Fatalf("mode %d duration = %v, want DurationPermanent", i+1, apply.Duration)
		}
		if !reflect.DeepEqual(apply.ContinuousEffects[0].AddKeywords, []game.Keyword{want}) {
			t.Fatalf("mode %d keywords = %v, want %v", i+1, apply.ContinuousEffects[0].AddKeywords, want)
		}
	}
}
