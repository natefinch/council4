package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
)

// TestLowerLeadingTappedAndAttackingToken verifies that a created token printing
// its entry modifier as a leading adjective ("create a tapped and attacking 1/1
// ... token", Pugnacious Pugilist / Ghalta and Mavren) lowers identically to the
// trailing "that's tapped and attacking" relative-clause form, emitting a
// CreateToken whose EntryTapped and EntryAttacking are both set.
func TestLowerLeadingTappedAndAttackingToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Leading Attack",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		Colors:     []string{"R"},
		OracleText: "Create a tapped and attacking 1/1 red Warrior creature token.",
	})
	create := createTokenPrimitive(t, face)
	if !create.EntryTapped {
		t.Fatal("EntryTapped = false, want true")
	}
	if !create.EntryAttacking {
		t.Fatal("EntryAttacking = false, want true")
	}
}

// TestLowerLeadingTappedAndAttackingTokenMatchesTrailing verifies the leading and
// trailing orderings of the "tapped and attacking" entry modifier produce the
// same token definition, so the leading form is a pure reconstruction broadening
// rather than a new shape.
func TestLowerLeadingTappedAndAttackingTokenMatchesTrailing(t *testing.T) {
	t.Parallel()
	leading := createTokenPrimitive(t, lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Leading Attack Order",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		Colors:     []string{"R"},
		OracleText: "Create a tapped and attacking 1/1 red Warrior creature token.",
	}))
	trailing := createTokenPrimitive(t, lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Trailing Attack Order",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		Colors:     []string{"R"},
		OracleText: "Create a 1/1 red Warrior creature token that's tapped and attacking.",
	}))
	if leading.EntryTapped != trailing.EntryTapped ||
		leading.EntryAttacking != trailing.EntryAttacking {
		t.Fatalf("leading entry %+v/%+v, trailing entry %+v/%+v",
			leading.EntryTapped, leading.EntryAttacking,
			trailing.EntryTapped, trailing.EntryAttacking)
	}
}

// TestLowerMultiColorCreatureToken verifies that a created creature token with
// three colors ("create a 1/1 red, green, and white Sand Warrior creature token",
// Sand Scout) lowers to a token definition carrying every printed color, the
// multi-color broadening of the single-color creature-token form.
func TestLowerMultiColorCreatureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tricolor Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		Colors:     []string{"R", "G", "W"},
		OracleText: "Create a 1/1 red, green, and white Sand Warrior creature token.",
	})
	create := createTokenPrimitive(t, face)
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	want := []color.Color{color.Red, color.Green, color.White}
	if len(def.Colors) != len(want) {
		t.Fatalf("token colors = %v, want %v", def.Colors, want)
	}
	for i, c := range want {
		if def.Colors[i] != c {
			t.Fatalf("token colors = %v, want %v", def.Colors, want)
		}
	}
}
