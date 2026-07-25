package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestRenderEventKindCoverage asserts that every non-unknown game.EventKind is
// classified exactly once: either renderEventKind returns a non-error rendering,
// or the kind is recorded in unrenderedEventKinds and renderEventKind fails
// closed. A newly added EventKind that is neither rendered nor recorded here, or
// a mapping that drifts from the recorded set, fails this test by name.
//
// EventKind is the one enum that still needs a test like this, because
// unrenderedEventKinds is a deliberate hand-maintained exclusion rather than a
// gap in a table. Every other enum's coverage now follows from the generated
// tables and is guarded by TestGeneratedFileIsCurrent in cardgen/cmd/genrender.
func TestRenderEventKindCoverage(t *testing.T) {
	t.Parallel()
	for kind := game.EventUnknown + 1; int(kind) < game.EventKindCount; kind++ {
		_, err := renderEventKind(kind)
		intentionallyUnrendered := unrenderedEventKinds[kind]
		switch {
		case err == nil && intentionallyUnrendered:
			t.Errorf("event kind %d is now rendered; remove it from unrenderedEventKinds", int(kind))
		case err != nil && !intentionallyUnrendered:
			t.Errorf("event kind %d has no renderer and is not recorded as intentionally unrendered: %v", int(kind), err)
		default:
			// Classified consistently: rendered-and-expected, or
			// unrendered-and-recorded. No drift.
		}
	}
}

// TestRenderEventKindRejectsUnknown pins the fail-closed behavior for the unset
// sentinel. game.EventUnknown is the zero EventKind and means "no event was
// specified", so a trigger pattern carrying it is malformed and must fail the
// render rather than emit game.EventUnknown into generated card source.
func TestRenderEventKindRejectsUnknown(t *testing.T) {
	t.Parallel()
	if _, err := renderEventKind(game.EventUnknown); err == nil {
		t.Fatal("expected an error for the unset event kind sentinel")
	}
}

// declared constants fails rather than emitting an unparseable literal.
func TestEnumLiteralRejectsUndeclaredValue(t *testing.T) {
	t.Parallel()
	if _, err := renderDuration(game.EffectDuration(250)); err == nil {
		t.Fatal("expected an error for an undeclared effect duration")
	}
}

// TestEnumLiteralNonZeroRejectsZero asserts the types whose zero value means
// "unset" refuse to emit it, while a type whose zero value is a real value still
// renders. Losing that distinction would emit fields asserting something the
// CardDef does not say.
func TestEnumLiteralNonZeroRejectsZero(t *testing.T) {
	t.Parallel()
	if _, err := renderZone(zone.None); err == nil {
		t.Fatal("expected an error for the unset zone")
	}
	if _, err := renderStep(game.StepNone); err == nil {
		t.Fatal("expected an error for the unset step")
	}
	literal, err := renderDuration(game.DurationPermanent)
	if err != nil {
		t.Fatalf("renderDuration(DurationPermanent): %v", err)
	}
	if literal != "game.DurationPermanent" {
		t.Errorf("renderDuration(DurationPermanent) = %q, want game.DurationPermanent", literal)
	}
}

// TestRenderLiteralsCoverFormerlyMissingConstants pins the constants that the
// hand-written switches did not map. Each is a declared value that a lowering
// may legitimately produce, and each previously failed to render, which dropped
// the card rather than reporting a diagnostic.
func TestRenderLiteralsCoverFormerlyMissingConstants(t *testing.T) {
	t.Parallel()
	steps := []game.Step{
		game.StepUntap,
		game.StepDeclareBlockers,
		game.StepFirstStrikeDamage,
		game.StepCombatDamage,
		game.StepCleanup,
	}
	for _, step := range steps {
		if _, err := renderStep(step); err != nil {
			t.Errorf("renderStep(%d): %v", int(step), err)
		}
	}
	if _, err := renderZone(zone.Stack); err != nil {
		t.Errorf("renderZone(zone.Stack): %v", err)
	}
	if _, err := renderDuration(game.DurationNextTime); err != nil {
		t.Errorf("renderDuration(DurationNextTime): %v", err)
	}
	if _, err := renderResolutionChoiceKind(game.ResolutionChoiceCardName); err != nil {
		t.Errorf("renderResolutionChoiceKind(ResolutionChoiceCardName): %v", err)
	}
}

// TestBitmaskLiteralCombinesFlags asserts a multi-flag value renders as an OR in
// ascending flag order, and that an undeclared bit fails.
func TestBitmaskLiteralCombinesFlags(t *testing.T) {
	t.Parallel()
	literal, err := renderDamageRecipient(game.DamageRecipientPlayer | game.DamageRecipientPermanent)
	if err != nil {
		t.Fatalf("renderDamageRecipient: %v", err)
	}
	const want = "game.DamageRecipientPlayer | game.DamageRecipientPermanent"
	if literal != want {
		t.Errorf("renderDamageRecipient = %q, want %q", literal, want)
	}
	if _, err := renderDamageRecipient(game.DamageRecipientKind(1 << 6)); err == nil {
		t.Fatal("expected an error for an undeclared damage recipient flag")
	}
}
