package cardgen

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/natefinch/council4/mtg/game"
)

// This file holds the assertion helpers card-support tests should use. See
// README.md ("Testing convention") for when to reach for each one.
//
// The convention it replaces matched substrings against the renderer's Go
// source output. That coupled every card test to renderer formatting: a
// renderer refactor that altered no compiled card still forced roughly 120
// assertion rewrites. It was also imprecise, because a substring matches
// anywhere in the file and happily passes against a different ability than the
// one under test.
//
// A CardDef path assertion has neither problem. It names one field of one
// value, so it cannot match the wrong ability, and it is immune to every
// formatting, spelling, and always-emit decision the renderer makes.

// cardDefPaths renders a compiled card as a flat, deterministic list of
// "path = value" lines, one per non-zero leaf field.
//
// Flat paths rather than an indented tree because an assertion has to be
// greppable and self-contained: "Abilities[0].Trigger.Pattern.ExcludeSelf =
// true" carries its own context, where a bare "ExcludeSelf: true" line does not
// say which ability it belongs to.
//
// Zero-valued fields are omitted, so a path's presence means "this field was
// deliberately set" and its absence means "this field holds its zero value".
// Both are assertable.
func cardDefPaths(t *testing.T, card *ScryfallCard) []string {
	t.Helper()
	defs, diagnostics, err := CompileCardDefs(card)
	if err != nil {
		t.Fatalf("compile %s: %v", card.Name, err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("compile %s produced diagnostics: %#v", card.Name, diagnostics)
	}
	if len(defs) == 0 {
		t.Fatalf("compile %s produced no CardDef", card.Name)
	}
	var paths []string
	for i, def := range defs {
		// A single-CardDef card, the overwhelming majority, gets unprefixed
		// paths so an assertion reads the way the Go type does. Reversible
		// cards yield one CardDef per face and need the index.
		prefix := "CardDef"
		if len(defs) > 1 {
			prefix = fmt.Sprintf("CardDef[%d]", i)
		}
		paths = appendPaths(paths, prefix, reflect.ValueOf(def))
	}
	return paths
}

// assertCardPaths compiles card and requires every wanted path to be present.
//
// Each want is matched as a substring of a single path line, so a test can
// assert the full path or just its distinctive tail. Matching within one line
// keeps the precision that source-text matching lacked: a want can never be
// satisfied by text from two different abilities.
func assertCardPaths(t *testing.T, card *ScryfallCard, want ...string) {
	t.Helper()
	paths := cardDefPaths(t, card)
	missing := missingPaths(paths, want)
	if len(missing) == 0 {
		return
	}
	t.Fatalf("%s: CardDef is missing %d expected path(s):\n  %s\n\ncompiled CardDef:\n%s",
		card.Name, len(missing), strings.Join(missing, "\n  "), strings.Join(paths, "\n"))
}

// assertCardPathsAbsent requires that no compiled path matches any of the given
// substrings.
//
// Absence assertions matter for fail-closed properties: "this card does not
// carry a target", "this trigger has no intervening if". Under the old
// convention those were written as strings.Contains negations against rendered
// source, where an unrelated formatting change could make them pass vacuously.
func assertCardPathsAbsent(t *testing.T, card *ScryfallCard, unwanted ...string) {
	t.Helper()
	found := forbiddenMatches(cardDefPaths(t, card), unwanted)
	if len(found) > 0 {
		t.Fatalf("%s: CardDef carries paths that must be absent:\n  %s", card.Name, strings.Join(found, "\n  "))
	}
}

// assertCardUnsupported requires that a card fails to compile, and that at least
// one diagnostic mentions each given substring.
//
// Fail-closed behavior is a feature of this compiler, so the cards it must
// refuse deserve tests as much as the cards it accepts.
func assertCardUnsupported(t *testing.T, card *ScryfallCard, wantDiagnostics ...string) {
	t.Helper()
	defs, diagnostics, err := CompileCardDefs(card)
	if err != nil {
		t.Fatalf("compile %s: %v", card.Name, err)
	}
	if len(diagnostics) == 0 {
		t.Fatalf("%s compiled into %d CardDef(s), want it refused", card.Name, len(defs))
	}
	var summaries []string
	for _, d := range diagnostics {
		summaries = append(summaries, d.Summary+": "+d.Detail)
	}
	for _, want := range wantDiagnostics {
		if !anyContains(summaries, want) {
			t.Fatalf("%s: no diagnostic mentions %q; got:\n  %s", card.Name, want, strings.Join(summaries, "\n  "))
		}
	}
}

// missingPaths returns the wants no path line satisfies. It is separated from
// the assertion so the matching rule itself is directly testable.
func missingPaths(paths, want []string) []string {
	var missing []string
	for _, w := range want {
		if !anyContains(paths, w) {
			missing = append(missing, w)
		}
	}
	return missing
}

// forbiddenMatches returns a description of every path line that matches an
// unwanted substring.
func forbiddenMatches(paths, unwanted []string) []string {
	var found []string
	for _, u := range unwanted {
		for _, p := range paths {
			if strings.Contains(p, u) {
				found = append(found, fmt.Sprintf("%q matched %q", u, p))
			}
		}
	}
	return found
}

func anyContains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if strings.Contains(h, needle) {
			return true
		}
	}
	return false
}

var (
	pathStringerType = reflect.TypeFor[fmt.Stringer]()
	abilityType      = reflect.TypeFor[game.Ability]()
)

// appendPaths walks v and appends one "path = value" line per non-zero leaf.
//
// Reflection is confined to test helpers deliberately. Production rendering is
// generated from the type graph precisely because reflect cannot enumerate
// constants, but a test dumper needs no constant names beyond what fmt.Stringer
// already supplies, and generating a second walker for it would cost far more
// than it returns.
func appendPaths(out []string, path string, v reflect.Value) []string {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return out
		}
		return appendPaths(out, path, v.Elem())
	case reflect.Interface:
		if v.IsNil() {
			return out
		}
		// Name the concrete type, because which implementation an interface
		// field holds is usually the most important fact about it: whether an
		// ability is triggered or activated, which primitive an effect runs.
		// An interface's payload is not addressable, so copy it into an
		// addressable cell first: readable needs an address to reach the
		// unexported fields below.
		concrete := v.Elem()
		if !concrete.CanAddr() && concrete.CanInterface() {
			cell := reflect.New(concrete.Type()).Elem()
			cell.Set(concrete)
			concrete = cell
		}
		return appendPaths(out, fmt.Sprintf("%s.(%s)", path, typeName(concrete)), concrete)
	case reflect.Struct:
		if name, ok := pathStringerName(v); ok {
			return append(out, path+" = "+name)
		}
		t := v.Type()
		before := len(out)
		for i := range v.NumField() {
			fv := readable(v.Field(i))
			if fv.IsZero() {
				continue
			}
			// Unexported fields are walked too. Several primitives keep their
			// whole payload in unexported state behind constructors, most
			// notably game.CreateToken's TokenSource: skipping those fields
			// dropped the entire token specification from the dump, so an
			// assertion about the token being a 1/1 Nightmare copy would have
			// passed against a card that created nothing of the sort. A dump
			// that silently omits content is worse than no dump at all.
			out = appendPaths(out, path+"."+t.Field(i).Name, fv)
		}
		if len(out) == before && !v.IsZero() {
			// A non-zero value that produced no paths would be invisible to
			// every assertion. Say so loudly rather than let a test pass
			// vacuously; TestCardDefPathsWalkAllReachableState fails on this.
			out = append(out, fmt.Sprintf("%s = <unwalkable %s>", path, t.String()))
		}
		return out
	case reflect.Slice, reflect.Array:
		for i := range v.Len() {
			out = appendPaths(out, fmt.Sprintf("%s[%d]", path, i), v.Index(i))
		}
		return out
	case reflect.Map:
		// Map iteration order is random, so collect and sort by rendered key
		// to keep the dump stable across runs.
		keys := v.MapKeys()
		slices.SortFunc(keys, func(a, b reflect.Value) int {
			return strings.Compare(leafString(a), leafString(b))
		})
		for _, k := range keys {
			elem := v.MapIndex(k)
			if elem.CanInterface() {
				cell := reflect.New(elem.Type()).Elem()
				cell.Set(elem)
				elem = cell
			}
			out = appendPaths(out, fmt.Sprintf("%s[%s]", path, leafString(k)), elem)
		}
		return out
	default:
		return append(out, path+" = "+leafString(v))
	}
}

// readable re-exports an unexported field so the rest of the dumper can treat it
// like any other value.
//
// Several primitives keep their entire payload in unexported state behind
// constructors: game.CreateToken's TokenSource holds the whole token
// specification, and game.ObjectReference holds the link it points at. Walking
// those fields is not optional, because a dump that omitted them would let an
// assertion about a 1/1 Nightmare token copy pass against a card that creates
// something else entirely.
//
// Without an address, reflect can still read an unexported field's structure but
// cannot turn it back into an interface, so every enum below it would print as a
// bare integer and every types.Card as a quoted string. Taking the address
// restores the constant spellings that make a failure message readable.
//
// This is test-only, and it only ever reads. Production code has no equivalent:
// the renderer emits these types through hand-written opaque emitters that use
// their exported accessors.
func readable(v reflect.Value) reflect.Value {
	if v.CanInterface() || !v.CanAddr() {
		return v
	}
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem() //nolint:gosec // test-only read of unexported state; see doc comment
}

// typeName is a value's type spelled the way Go source spells it, so a path
// reads like the code it describes.
func typeName(v reflect.Value) string {
	t := v.Type()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.PkgPath() == "" {
		return t.String()
	}
	return t.String()
}

// leafString renders a scalar, preferring the generated constant spelling so an
// enum reads as "game.EventPermanentEnteredBattlefield" rather than "3".
//
// Naming constants is exactly what reflection cannot do and what the generator
// can, so the spelling comes from the generated table and only the walk is
// reflective.
func leafString(v reflect.Value) string {
	if v.CanInterface() {
		if name, ok := enumSpelling(v.Interface()); ok && name != "" {
			return name
		}
	}
	if name, ok := pathStringerName(v); ok {
		return name
	}
	if v.Kind() == reflect.String {
		return strconv.Quote(v.String())
	}
	if v.CanInterface() {
		return fmt.Sprintf("%v", v.Interface())
	}
	// An unexported field cannot be turned back into an interface, so it
	// cannot consult enumSpelling and prints numerically. That is a readable
	// enough compromise for internal state, which is always reached through
	// an exported constructor whose arguments the surrounding paths show.
	switch v.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// pathStringerName returns a value's fmt.Stringer rendering. Unlike the golden
// harness's dumper this also applies to structs, because the opaque reference
// types (game.ObjectReference and friends) have unexported state that reflection
// cannot usefully walk but do print themselves meaningfully.
func pathStringerName(v reflect.Value) (string, bool) {
	if !v.CanInterface() || !v.Type().Implements(pathStringerType) {
		return "", false
	}
	// An ability is a Stringer too, but its whole structure is what a test
	// wants to see, so keep walking it.
	if v.Type().Implements(abilityType) {
		return "", false
	}
	s, ok := v.Interface().(fmt.Stringer)
	if !ok {
		return "", false
	}
	return s.String(), true
}

// pathValue returns the value of the single path line matching want, failing if
// no line or more than one line matches.
//
// Some properties are relations between two fields rather than a fixed value: a
// published result key must be the one the later instruction gates on, whatever
// key the lowering chose. Asserting the literal key would pin an incidental
// naming choice; comparing the two occurrences pins the property that matters.
func pathValue(t *testing.T, paths []string, want string) string {
	t.Helper()
	var matches []string
	for _, p := range paths {
		if strings.Contains(p, want) {
			matches = append(matches, p)
		}
	}
	if len(matches) != 1 {
		t.Fatalf("want exactly one path containing %q, found %d:\n%s", want, len(matches), strings.Join(matches, "\n"))
	}
	_, value, found := strings.Cut(matches[0], " = ")
	if !found {
		t.Fatalf("path %q has no value", matches[0])
	}
	return value
}
