package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestSelectionPhraseMassRoundTrip renders typed permanent-group selections
// through the canonical selectionPhrase renderer and compares the result against
// the exact Oracle noun phrase the mass/all/each family reconstructs. These are
// the byte-exact forms exactMassGroupPhrase / exactMassEachGroupPhrase now verify
// the typed selection against, so a drift in the renderer surfaces here.
func TestSelectionPhraseMassRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		selection SelectionSyntax
		number    grammaticalNumber
		want      string
	}{
		{
			name:      "plain creatures",
			selection: SelectionSyntax{Kind: SelectionCreature, All: true, RequiredTypesAny: []CardType{CardTypeCreature}},
			number:    numberPlural,
			want:      "creatures",
		},
		{
			name:      "plain artifacts",
			selection: SelectionSyntax{Kind: SelectionArtifact, All: true, RequiredTypesAny: []CardType{CardTypeArtifact}},
			number:    numberPlural,
			want:      "artifacts",
		},
		{
			name:      "nonland permanents",
			selection: SelectionSyntax{Kind: SelectionPermanent, All: true, ExcludedTypes: []CardType{CardTypeLand}},
			number:    numberPlural,
			want:      "nonland permanents",
		},
		{
			name:      "nonbasic lands",
			selection: SelectionSyntax{Kind: SelectionLand, All: true, RequiredTypesAny: []CardType{CardTypeLand}, ExcludedSupertypes: []Supertype{SupertypeBasic}},
			number:    numberPlural,
			want:      "nonbasic lands",
		},
		{
			name:      "untapped creatures",
			selection: SelectionSyntax{Kind: SelectionCreature, All: true, Untapped: true, RequiredTypesAny: []CardType{CardTypeCreature}},
			number:    numberPlural,
			want:      "untapped creatures",
		},
		{
			name:      "other creatures",
			selection: SelectionSyntax{Kind: SelectionCreature, All: true, Other: true, RequiredTypesAny: []CardType{CardTypeCreature}},
			number:    numberPlural,
			want:      "other creatures",
		},
		{
			name:      "white creatures",
			selection: SelectionSyntax{Kind: SelectionCreature, All: true, RequiredTypesAny: []CardType{CardTypeCreature}, ColorsAny: []Color{ColorWhite}},
			number:    numberPlural,
			want:      "white creatures",
		},
		{
			name:      "creatures you control",
			selection: SelectionSyntax{Kind: SelectionCreature, All: true, Controller: SelectionControllerYou, RequiredTypesAny: []CardType{CardTypeCreature}},
			number:    numberPlural,
			want:      "creatures you control",
		},
		{
			name:      "creatures your opponents control",
			selection: SelectionSyntax{Kind: SelectionCreature, All: true, Controller: SelectionControllerOpponent, RequiredTypesAny: []CardType{CardTypeCreature}},
			number:    numberPlural,
			want:      "creatures your opponents control",
		},
		{
			name:      "type union creatures and lands",
			selection: SelectionSyntax{Kind: SelectionCreature, All: true, RequiredTypesAny: []CardType{CardTypeCreature, CardTypeLand}},
			number:    numberPlural,
			want:      "creatures and lands",
		},
		{
			name:      "type union three members",
			selection: SelectionSyntax{Kind: SelectionArtifact, All: true, RequiredTypesAny: []CardType{CardTypeArtifact, CardTypeCreature, CardTypeEnchantment}},
			number:    numberPlural,
			want:      "artifacts, creatures, and enchantments",
		},
		{
			name:      "artifacts with mana value",
			selection: SelectionSyntax{Kind: SelectionArtifact, All: true, MatchManaValue: true, RequiredTypesAny: []CardType{CardTypeArtifact}, ManaValue: compare.Int{Op: compare.LessOrEqual, Value: 3}},
			number:    numberPlural,
			want:      "artifacts with mana value 3 or less",
		},
		{
			name:      "permanents with mana value or greater",
			selection: SelectionSyntax{Kind: SelectionPermanent, All: true, MatchManaValue: true, ManaValue: compare.Int{Op: compare.GreaterOrEqual, Value: 2}},
			number:    numberPlural,
			want:      "permanents with mana value 2 or greater",
		},
		{
			name:      "each singular creature",
			selection: SelectionSyntax{Kind: SelectionCreature, All: true, RequiredTypesAny: []CardType{CardTypeCreature}},
			number:    numberSingular,
			want:      "creature",
		},
		{
			name:      "each nonland permanent with mana value",
			selection: SelectionSyntax{Kind: SelectionPermanent, All: true, MatchManaValue: true, ExcludedTypes: []CardType{CardTypeLand}, ManaValue: compare.Int{Op: compare.LessOrEqual, Value: 2}},
			number:    numberSingular,
			want:      "nonland permanent with mana value 2 or less",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := selectionPhrase(tc.selection, selectionPhraseOptions{Number: tc.number})
			if !ok {
				t.Fatalf("selectionPhrase(%s) ok = false, want true", tc.name)
			}
			if got != tc.want {
				t.Errorf("selectionPhrase(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

// TestSelectionPhraseRejectsUnrepresentable proves the renderer fails closed for
// selection qualifiers the permanent-group context cannot represent. These forms
// stay owned by their existing family validators; selectionPhrase returning
// ok=false lets the mass family fall back to text-shape validation for them
// rather than rendering a phrase that silently drops the qualifier.
func TestSelectionPhraseRejectsUnrepresentable(t *testing.T) {
	t.Parallel()
	base := func() SelectionSyntax {
		return SelectionSyntax{Kind: SelectionCreature, All: true, RequiredTypesAny: []CardType{CardTypeCreature}}
	}
	cases := []struct {
		name   string
		mutate func(*SelectionSyntax)
	}{
		{"subtype", func(s *SelectionSyntax) { s.SubtypesAny = []types.Sub{types.Sub("Goblin")} }},
		{"keyword", func(s *SelectionSyntax) { s.Keyword = KeywordFlying }},
		{"excluded keyword", func(s *SelectionSyntax) { s.ExcludedKeyword = KeywordFlying }},
		{"supertype", func(s *SelectionSyntax) { s.Supertypes = []Supertype{SupertypeBasic} }},
		{"token only", func(s *SelectionSyntax) { s.TokenOnly = true }},
		{"nontoken", func(s *SelectionSyntax) { s.NonToken = true }},
		{"colorless", func(s *SelectionSyntax) { s.Colorless = true }},
		{"multicolored", func(s *SelectionSyntax) { s.Multicolored = true }},
		{"another", func(s *SelectionSyntax) { s.Another = true }},
		{"power less than source", func(s *SelectionSyntax) { s.PowerLessThanSource = true }},
		{"two colors", func(s *SelectionSyntax) { s.ColorsAny = []Color{ColorWhite, ColorBlue} }},
		{"two excluded types", func(s *SelectionSyntax) {
			s.ExcludedTypes = []CardType{CardTypeLand, CardTypeArtifact}
		}},
		{"non-permanent kind", func(s *SelectionSyntax) {
			s.Kind = SelectionUnknown
			s.RequiredTypesAny = nil
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			selection := base()
			tc.mutate(&selection)
			if got, ok := selectionPhrase(selection, selectionPhraseOptions{Number: numberPlural}); ok {
				t.Errorf("selectionPhrase(%s) = %q, true; want fail closed", tc.name, got)
			}
		})
	}
}

// TestSelectionPhraseRejectsUnsupportedOptions proves the renderer fails closed
// for option contexts later stages own (a non-empty determiner, a zone noun, or
// a card noun), so a caller cannot get a permanent-group rendering for a context
// this stage does not implement.
func TestSelectionPhraseRejectsUnsupportedOptions(t *testing.T) {
	t.Parallel()
	selection := SelectionSyntax{Kind: SelectionCreature, All: true, RequiredTypesAny: []CardType{CardTypeCreature}}
	cases := []struct {
		name string
		opts selectionPhraseOptions
	}{
		{"zone noun", selectionPhraseOptions{Number: numberPlural, ZoneNoun: true}},
		{"card noun", selectionPhraseOptions{Number: numberPlural, CardNoun: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got, ok := selectionPhrase(selection, tc.opts); ok {
				t.Errorf("selectionPhrase(%s) = %q, true; want fail closed", tc.name, got)
			}
		})
	}
}

// TestSelectionPhraseVerifiesMassGroupGap proves the typed verification closes
// the soundness gap: a selection whose typed noun disagrees with the source
// phrase fails verification even though the text shape alone would accept it.
func TestSelectionPhraseVerifiesMassGroupGap(t *testing.T) {
	t.Parallel()
	// Text shape accepts the literal phrase "creatures", but the typed selection
	// describes an artifact group, so the typed verification must reject it.
	selection := SelectionSyntax{Kind: SelectionArtifact, All: true, RequiredTypesAny: []CardType{CardTypeArtifact}}
	if selectionPhraseVerifiesMassGroup(&selection, "creatures", numberPlural) {
		t.Errorf("selectionPhraseVerifiesMassGroup(artifact selection, %q) = true, want false", "creatures")
	}
	if !selectionPhraseVerifiesMassGroup(&selection, "artifacts", numberPlural) {
		t.Errorf("selectionPhraseVerifiesMassGroup(artifact selection, %q) = false, want true", "artifacts")
	}
}
