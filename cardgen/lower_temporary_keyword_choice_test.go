package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// assertUntilEndOfTurnChoiceModes verifies a modal AbilityContent grants exactly
// one of the listed keywords, with each mode applying a single until-end-of-turn
// keyword grant in the listed order.
func assertUntilEndOfTurnChoiceModes(t *testing.T, content game.AbilityContent, want []game.Keyword) {
	t.Helper()
	if content.MinModes != 1 || content.MaxModes != 1 {
		t.Fatalf("modes range = [%d,%d], want exactly one", content.MinModes, content.MaxModes)
	}
	if len(content.Modes) != len(want) {
		t.Fatalf("modes = %d, want %d", len(content.Modes), len(want))
	}
	for i, keyword := range want {
		mode := content.Modes[i]
		apply, ok := mode.Sequence[0].Primitive.(game.ApplyContinuous)
		if !ok {
			t.Fatalf("mode %d primitive = %T, want game.ApplyContinuous", i, mode.Sequence[0].Primitive)
		}
		if apply.Duration != game.DurationUntilEndOfTurn {
			t.Fatalf("mode %d duration = %v, want DurationUntilEndOfTurn", i, apply.Duration)
		}
		if !reflect.DeepEqual(apply.ContinuousEffects[0].AddKeywords, []game.Keyword{keyword}) {
			t.Fatalf("mode %d keywords = %v, want %v", i, apply.ContinuousEffects[0].AddKeywords, keyword)
		}
	}
}

// TestLowerTargetUntilEndOfTurnKeywordChoiceSpell proves a spell granting a
// target creature its controller's choice of several keywords until end of turn
// lowers to a single-shared-target modal content with one until-end-of-turn
// grant per keyword choice.
func TestLowerTargetUntilEndOfTurnKeywordChoiceSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Target Choice",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gains your choice of flying, first strike, or trample until end of turn.",
	})
	content := face.SpellAbility.Val
	if len(content.SharedTargets) != 1 {
		t.Fatalf("shared targets = %d, want 1", len(content.SharedTargets))
	}
	assertUntilEndOfTurnChoiceModes(t, content, []game.Keyword{game.Flying, game.FirstStrike, game.Trample})
}

// TestLowerSourceUntilEndOfTurnKeywordChoiceActivated proves an activated ability
// granting the source its controller's choice of keywords until end of turn
// lowers to modal content anchored on the source permanent.
func TestLowerSourceUntilEndOfTurnKeywordChoiceActivated(t *testing.T) {
	t.Parallel()
	power, toughness := "1", "1"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Source Choice",
		Layout:     "normal",
		TypeLine:   "Creature — Human Assassin",
		ManaCost:   "{B}",
		OracleText: "{1}: This creature gains your choice of flying, deathtouch, or lifelink until end of turn.",
		Power:      &power,
		Toughness:  &toughness,
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	content := face.ActivatedAbilities[0].Content
	if len(content.SharedTargets) != 0 {
		t.Fatalf("shared targets = %d, want 0 (source grant)", len(content.SharedTargets))
	}
	assertUntilEndOfTurnChoiceModes(t, content, []game.Keyword{game.Flying, game.Deathtouch, game.Lifelink})
}

// TestLowerUntilEndOfTurnKeywordConjunctionStillSingleMode proves the choice
// path does not capture conjunctions: granting several keywords joined by "and"
// until end of turn still lowers to one mode adding every keyword at once.
func TestLowerUntilEndOfTurnKeywordConjunctionStillSingleMode(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Conjunction",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gains flying and first strike until end of turn.",
	})
	content := face.SpellAbility.Val
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %d, want 1 (conjunction grants all at once)", len(content.Modes))
	}
	apply, ok := content.Modes[0].Sequence[0].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyContinuous", content.Modes[0].Sequence[0].Primitive)
	}
	if apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want DurationUntilEndOfTurn", apply.Duration)
	}
	want := []game.Keyword{game.Flying, game.FirstStrike}
	if !reflect.DeepEqual(apply.ContinuousEffects[0].AddKeywords, want) {
		t.Fatalf("keywords = %v, want %v", apply.ContinuousEffects[0].AddKeywords, want)
	}
}

// TestLowerGroupUntilEndOfTurnKeywordChoiceFailsClosed proves a group grant of a
// keyword choice until end of turn is not yet supported and fails closed rather
// than mis-lowering as a conjunction granting every keyword.
func TestLowerGroupUntilEndOfTurnKeywordChoiceFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Group Choice",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Creatures you control gain your choice of flying, first strike, or trample until end of turn.",
	})
	if face.SpellAbility.Exists {
		t.Fatalf("expected no spell ability for unsupported group choice, got %+v", face.SpellAbility.Val)
	}
}
