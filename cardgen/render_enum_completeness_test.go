package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// unrenderedEventKinds lists the game.EventKind values that renderEventKind
// deliberately does not map to generated source. These events are consumed only
// by runtime machinery (replacement effects, copies, phasing, reveals) and are
// never embedded as literals in generated card definitions, so
// fail-closed rendering is correct for them.
//
// This set exists so EventKind drift fails loudly: adding a new EventKind, or
// adding/removing a renderEventKind case, makes TestRenderEventKindCoverage
// fail until either a rendering is added or the kind is recorded here as
// intentionally unrendered.
var unrenderedEventKinds = map[game.EventKind]bool{
	game.EventSpellResolved:      true,
	game.EventDamagePrevented:    true,
	game.EventDestroyReplaced:    true,
	game.EventCardRevealed:       true,
	game.EventSpellCopied:        true,
	game.EventPermanentPhasedOut: true,
	game.EventPermanentPhasedIn:  true,
	// These dungeon/initiative events are produced and consumed only by runtime
	// machinery; no card text lowers to a trigger on them.
	game.EventVenturedIntoDungeon: true,
	game.EventTookInitiative:      true,
}

// TestRenderEventKindCoverage asserts that every non-unknown game.EventKind is
// classified exactly once: either renderEventKind returns a non-error rendering,
// or the kind is recorded in unrenderedEventKinds and renderEventKind fails
// closed. A newly added EventKind that is neither rendered nor recorded here, or
// a mapping that drifts from the recorded set, fails this test by name.
func TestRenderEventKindCoverage(t *testing.T) {
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

// TestRenderDurationCoverage asserts renderDuration maps every declared
// game.EffectDuration value, since durations are emitted directly into generated
// source. EffectDuration has no count sentinel, so the closed list below is the
// authoritative set; adding a duration without a rendering fails here.
func TestRenderDurationCoverage(t *testing.T) {
	durations := []game.EffectDuration{
		game.DurationPermanent,
		game.DurationUntilEndOfTurn,
		game.DurationUntilYourNextTurn,
		game.DurationThisTurn,
		game.DurationUntilEndOfYourNextTurn,
		game.DurationUntilYourNextEndStep,
		game.DurationForAsLongAsSourceOnBattlefield,
		game.DurationForAsLongAsYouControlSource,
		game.DurationForAsLongAsControlledCreatureEnchanted,
	}
	for _, d := range durations {
		if _, err := renderDuration(d); err != nil {
			t.Errorf("duration %d has no renderer: %v", int(d), err)
		}
	}
}
