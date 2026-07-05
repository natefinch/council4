package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseSingleAbility parses source expected to yield exactly one ability and
// returns it, so tests assert on typed syntax rather than source text.
func parseSingleAbility(t *testing.T, source string, context Context) Ability {
	t.Helper()
	document, diagnostics := Parse(source, context)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("source %q abilities = %d, want one", source, len(document.Abilities))
	}
	return document.Abilities[0]
}

func TestEmitConditionBoundaryIntroKinds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
		want   ConditionIntroKind
	}{
		{"if", "When this creature enters, if you control a Mountain, draw a card.", ConditionIntroIf},
		{"unless", "When this creature enters, sacrifice it unless you pay {2}.", ConditionIntroUnless},
		{"only if", "{T}: Draw a card. Activate only if you have 10 or more life.", ConditionIntroOnlyIf},
		{"as long as", "As long as you control a Mountain, this creature gets +1/+1.", ConditionIntroAsLongAs},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ability := parseSingleAbility(t, tc.source, Context{})
			var kinds []ConditionIntroKind
			for _, boundary := range ability.ConditionBoundaries {
				kinds = append(kinds, boundary.Kind)
			}
			if len(kinds) != 1 || kinds[0] != tc.want {
				t.Fatalf("boundary kinds = %#v, want exactly [%v]", kinds, tc.want)
			}
		})
	}
}

func TestEmitConditionBoundaryActivationKeyword(t *testing.T) {
	t.Parallel()

	// "Activate only if ..." records the preceding "Activate" keyword span so
	// lowering accounts for its consumed source without inspecting spelling.
	ability := parseSingleAbility(t,
		"{T}: Draw a card. Activate only if you have 10 or more life.", Context{})
	boundary := onlyIfBoundary(t, &ability)
	if boundary.ActivationKeyword == (shared.Span{}) {
		t.Fatal("expected an ActivationKeyword span for \"Activate only if\"")
	}
	got := ability.Text[boundary.ActivationKeyword.Start.Offset:boundary.ActivationKeyword.End.Offset]
	if got != "Activate" {
		t.Fatalf("ActivationKeyword span text = %q, want %q", got, "Activate")
	}

	// Near miss: "Activate this ability only if ..." is not the bare
	// "Activate only if" form, so no keyword span is recorded.
	nearMiss := parseSingleAbility(t,
		"{T}: Draw a card. Activate this ability only if you have 10 or more life.", Context{})
	if b := onlyIfBoundary(t, &nearMiss); b.ActivationKeyword != (shared.Span{}) {
		t.Fatalf("near miss recorded an ActivationKeyword span = %#v", b.ActivationKeyword)
	}
}

func TestEmitConditionBoundaryDurationSkip(t *testing.T) {
	t.Parallel()

	// A source-duration "as long as" must be skipped: compileDuration already
	// captures it, so it is not also a standalone condition.
	duration := parseSingleAbility(t,
		"This creature gets +1/+1 for as long as you control a Mountain.", Context{})
	if b := asLongAsBoundary(t, &duration); !b.DurationSkip {
		t.Fatal("\"for as long as\" boundary DurationSkip = false, want true")
	}

	// A plain "as long as" condition is a real condition, not a duration skip.
	condition := parseSingleAbility(t,
		"As long as you control a Mountain, this creature gets +1/+1.", Context{})
	if b := asLongAsBoundary(t, &condition); b.DurationSkip {
		t.Fatal("standalone \"as long as\" condition wrongly marked DurationSkip")
	}
}

func TestEmitConditionBoundaryInterveningIf(t *testing.T) {
	t.Parallel()

	// A triggered ability's "if" immediately after the event comma is an
	// intervening-if.
	triggered := parseSingleAbility(t,
		"When this creature enters, if you control a Mountain, draw a card.", Context{})
	if b := ifBoundary(t, &triggered); !b.Intervening {
		t.Fatal("triggered intervening-if boundary Intervening = false, want true")
	}

	// The same introducer in a non-triggered context is never intervening.
	static := parseSingleAbility(t,
		"As long as you control a Mountain, this creature gets +1/+1.", Context{})
	for _, b := range static.ConditionBoundaries {
		if b.Intervening {
			t.Fatalf("non-triggered boundary %#v wrongly marked Intervening", b)
		}
	}
}

// TestEmitConditionBoundaryInterveningWhile proves a triggered "while <state>"
// clause inside the event phrase is an intervening condition (Regal Behemoth's
// "Whenever you tap a land for mana while you're the monarch, ..."), while the
// combat event rider "attacks while saddled" is left to the event parser and
// never becomes a condition boundary.
func TestEmitConditionBoundaryInterveningWhile(t *testing.T) {
	t.Parallel()

	triggered := parseSingleAbility(t,
		"Whenever you tap a land for mana while you're the monarch, add {C}.", Context{})
	boundary := boundaryOfKind(t, &triggered, ConditionIntroAsLongAs)
	if !boundary.Intervening {
		t.Fatal("triggered \"while\" boundary Intervening = false, want true")
	}

	saddled := parseSingleAbility(t,
		"Whenever this creature attacks while saddled, draw a card.", Context{})
	for _, b := range saddled.ConditionBoundaries {
		t.Fatalf("\"while saddled\" event rider wrongly emitted a condition boundary: %#v", b)
	}
}

func TestEmitConditionBoundaryIfAbleExcluded(t *testing.T) {
	t.Parallel()
	// "attacks each combat if able" is captured as a structural restriction; the
	// trailing "if able" must not also become a standalone condition boundary.
	ability := parseSingleAbility(t,
		"When this creature enters, it attacks each combat if able.", Context{})
	for _, b := range ability.ConditionBoundaries {
		t.Fatalf("\"if able\" restriction wrongly emitted a condition boundary: %#v", b)
	}
}

func TestEmitOptionalYouMay(t *testing.T) {
	t.Parallel()

	// A triggered "you may" body is optional.
	optional := parseSingleAbility(t,
		"When this creature enters, you may draw a card.", Context{})
	if !optional.Optional {
		t.Fatal("triggered \"you may\" ability Optional = false, want true")
	}
	got := optional.Text[optional.OptionalSpan.Start.Offset:optional.OptionalSpan.End.Offset]
	if got != "you may" {
		t.Fatalf("OptionalSpan text = %q, want %q", got, "you may")
	}

	// Near miss: "you instead" is not "you may".
	nearMiss := parseSingleAbility(t,
		"When this creature enters, you instead draw two cards.", Context{})
	if nearMiss.Optional {
		t.Fatal("\"you instead\" wrongly marked Optional")
	}

	// Optionality is gated to triggered abilities: an activated "You may" body is
	// not treated as an ability-level optional flag here.
	activated := parseSingleAbility(t,
		"{T}: You may draw a card.", Context{})
	if activated.Optional {
		t.Fatal("activated ability wrongly marked Optional")
	}
}

func onlyIfBoundary(t *testing.T, ability *Ability) ConditionBoundary {
	t.Helper()
	return boundaryOfKind(t, ability, ConditionIntroOnlyIf)
}

func asLongAsBoundary(t *testing.T, ability *Ability) ConditionBoundary {
	t.Helper()
	return boundaryOfKind(t, ability, ConditionIntroAsLongAs)
}

func ifBoundary(t *testing.T, ability *Ability) ConditionBoundary {
	t.Helper()
	return boundaryOfKind(t, ability, ConditionIntroIf)
}

func boundaryOfKind(t *testing.T, ability *Ability, kind ConditionIntroKind) ConditionBoundary {
	t.Helper()
	for _, boundary := range ability.ConditionBoundaries {
		if boundary.Kind == kind {
			return boundary
		}
	}
	t.Fatalf("no condition boundary of kind %v in %#v", kind, ability.ConditionBoundaries)
	return ConditionBoundary{}
}
