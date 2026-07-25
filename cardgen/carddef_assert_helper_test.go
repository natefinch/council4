package cardgen

import (
	"slices"
	"strings"
	"testing"
)

// TestCardDefPathsNameEveryEnumByConstant is the property that makes a CardDef
// dump usable as a failure message. Reflection can walk the value but cannot
// name a constant, so without the generated enumSpelling table a failing
// assertion would read "Trigger.Pattern.Event = 3" and force the reader to go
// look up which event that is.
//
// It also guards the wiring: if enumSpelling were dropped from the generated
// file, or the dumper stopped consulting it, every enum in every card test's
// failure output would silently regress to a bare integer.
func TestCardDefPathsNameEveryEnumByConstant(t *testing.T) {
	t.Parallel()
	paths := cardDefPaths(t, kodamaEastTreeCard())
	for _, want := range []string{
		"Trigger.Type = game.TriggerWhenever",
		"Trigger.Pattern.Event = game.EventPermanentEnteredBattlefield",
		"Trigger.Pattern.Controller = game.TriggerControllerYou",
		"SourceZone = zone.Hand",
		"Filter.RequiredTypesAny[0] = types.Artifact",
		"Kind = game.Reach",
	} {
		if !anyContains(paths, want) {
			t.Errorf("dump has no path containing %q:\n%s", want, strings.Join(paths, "\n"))
		}
	}
}

// TestCardDefPathsQualifyInterfaceFields proves the dump records which
// implementation an interface field holds. Which primitive an effect runs is
// usually the most important fact about it, and a path that omitted it would let
// an assertion pass against a completely different effect.
func TestCardDefPathsQualifyInterfaceFields(t *testing.T) {
	t.Parallel()
	paths := cardDefPaths(t, kodamaEastTreeCard())
	if !anyContains(paths, "Primitive.(game.ChooseFromZone).") {
		t.Fatalf("dump does not name the concrete primitive type:\n%s", strings.Join(paths, "\n"))
	}
}

// TestCardDefPathsOmitZeroFields pins the rule that gives absence assertions
// meaning: a path appears only when its field was deliberately set. If zero
// fields were dumped, assertCardPathsAbsent would match every field of every
// type and could never pass.
func TestCardDefPathsOmitZeroFields(t *testing.T) {
	t.Parallel()
	paths := cardDefPaths(t, kodamaEastTreeCard())
	// Kodama's trigger is not restricted to a single controller's permanents
	// entering from a particular zone, so the zone-matching flags are unset.
	for _, unwanted := range []string{"Trigger.Pattern.MatchFromZone", "Trigger.Pattern.MatchToZone"} {
		if anyContains(paths, unwanted) {
			t.Errorf("dump contains zero-valued path %q:\n%s", unwanted, strings.Join(paths, "\n"))
		}
	}
}

// TestCardDefPathsAreDeterministic guards against unstable ordering. The walk
// visits struct fields in declaration order and slices in index order, but a map
// field would iterate randomly, so map keys are sorted. A dump that reordered
// between runs would make golden comparisons and diffs useless.
func TestCardDefPathsAreDeterministic(t *testing.T) {
	t.Parallel()
	first := cardDefPaths(t, kodamaEastTreeCard())
	second := cardDefPaths(t, kodamaEastTreeCard())
	if !slices.Equal(first, second) {
		t.Fatal("cardDefPaths is not deterministic across two compilations")
	}
}

// TestMissingPathsReportsUnmatchedWants proves the matcher actually rejects. A
// substring matcher that silently matched everything would turn every converted
// card test into a no-op, which is a worse failure than the fragility the
// convention replaces.
func TestMissingPathsReportsUnmatchedWants(t *testing.T) {
	t.Parallel()
	paths := []string{"CardDef.CardFace.Name = \"Shock\"", "CardDef.CardFace.Types[0] = types.Instant"}
	got := missingPaths(paths, []string{"Types[0] = types.Instant", "Types[0] = types.Sorcery"})
	want := []string{"Types[0] = types.Sorcery"}
	if !slices.Equal(got, want) {
		t.Fatalf("missingPaths = %v, want %v", got, want)
	}
}

// TestForbiddenMatchesReportsEachHit proves an absence assertion reports which
// path violated it, so a failure says what the card actually carries rather than
// only that something was wrong.
func TestForbiddenMatchesReportsEachHit(t *testing.T) {
	t.Parallel()
	paths := []string{"CardDef.CardFace.TriggeredAbilities[0].Trigger.Pattern.ExcludeSelf = true"}
	got := forbiddenMatches(paths, []string{"ExcludeSelf", "MatchFromZone"})
	if len(got) != 1 || !strings.Contains(got[0], "ExcludeSelf") {
		t.Fatalf("forbiddenMatches = %v, want one ExcludeSelf hit", got)
	}
}

// TestCardDefPathsWalkAllReachableState is the guard that keeps the dump
// honest. Several primitives keep their whole payload in unexported state
// behind constructors, and the first version of this dumper skipped unexported
// fields entirely: game.CreateToken produced no paths at all, so an assertion
// about the token it creates would have passed against a card that creates
// something completely different.
//
// A non-zero value the walk cannot describe is marked "<unwalkable>" rather
// than dropped, and this fails on the marker across every golden card. Silence
// is the one failure mode a test helper must never have.
func TestCardDefPathsWalkAllReachableState(t *testing.T) {
	t.Parallel()
	for _, gc := range goldenCards {
		t.Run(gc.Key, func(t *testing.T) {
			t.Parallel()
			for _, path := range cardDefPaths(t, gc.Card) {
				if strings.Contains(path, "<unwalkable") {
					t.Errorf("%s: %s", gc.Rationale, path)
				}
			}
		})
	}
}

// TestCardDefPathsDescribeOpaquePrimitiveState pins the specific regression
// above with the card that exposed it. Nightmare Shepherd's payoff is a token
// that copies the dead creature "except it's 1/1 and it's a Nightmare", and
// every one of those facts lives inside game.CreateToken's unexported
// TokenSource.
func TestCardDefPathsDescribeOpaquePrimitiveState(t *testing.T) {
	t.Parallel()
	paths := cardDefPaths(t, nightmareShepherdCard())
	for _, want := range []string{
		"Primitive.(game.CreateToken).Source.copy.Source = game.TokenCopySourceObject",
		"Primitive.(game.CreateToken).Source.copy.Object.kind = game.ObjectReferenceLinkedObject",
		"Primitive.(game.CreateToken).Source.copy.SetPower.Val.Value = 1",
		"Primitive.(game.CreateToken).Source.copy.AddSubtypes[0] = types.Nightmare",
	} {
		if !anyContains(paths, want) {
			t.Errorf("dump has no path containing %q:\n%s", want, strings.Join(paths, "\n"))
		}
	}
}
