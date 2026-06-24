package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestSelectionPhraseEachDeterminerRoundTrip renders typed group recipient
// selections through the canonical selectionPhrase renderer with the "each"
// determiner and compares the result against the exact Oracle recipient phrase
// the group damage family reconstructs. These are the byte-exact forms
// exactGroupDamagePermanentRecipientText now verifies the typed selection
// against, so a drift in the renderer surfaces here.
func TestSelectionPhraseEachDeterminerRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		selection SelectionSyntax
		want      string
	}{
		{
			name:      "plain creature",
			selection: SelectionSyntax{Kind: SelectionCreature},
			want:      "each creature",
		},
		{
			name:      "creature you control",
			selection: SelectionSyntax{Kind: SelectionCreature, Controller: SelectionControllerYou},
			want:      "each creature you control",
		},
		{
			name:      "creature your opponents control",
			selection: SelectionSyntax{Kind: SelectionCreature, Controller: SelectionControllerOpponent},
			want:      "each creature your opponents control",
		},
		{
			name:      "other attacking creature you control",
			selection: SelectionSyntax{Kind: SelectionCreature, Other: true, Attacking: true, Controller: SelectionControllerYou},
			want:      "each other attacking creature you control",
		},
		{
			name:      "white creature",
			selection: SelectionSyntax{Kind: SelectionCreature, ColorsAny: []Color{ColorWhite}},
			want:      "each white creature",
		},
		{
			name:      "noncreature artifact",
			selection: SelectionSyntax{Kind: SelectionArtifact, ExcludedTypes: []CardType{CardTypeCreature}},
			want:      "each noncreature artifact",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := selectionPhrase(tc.selection, selectionPhraseOptions{
				Number:     numberSingular,
				Determiner: determinerEach,
			})
			if !ok {
				t.Fatalf("selectionPhrase(%s) ok = false, want true", tc.name)
			}
			if got != tc.want {
				t.Errorf("selectionPhrase(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

// TestGroupDamageRecipientRoundTrip locks in that the migrated group damage
// recipient reconstruction still produces the exact Oracle recipient text for
// the permanent-group forms the renderer now cross-checks, and for the
// numeric/counter forms whose cross-check is intentionally skipped.
func TestGroupDamageRecipientRoundTrip(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		selection SelectionSyntax
		want      string
	}{
		{
			name:      "plain creature",
			selection: SelectionSyntax{Kind: SelectionCreature},
			want:      "each creature",
		},
		{
			name:      "creature you control",
			selection: SelectionSyntax{Kind: SelectionCreature, Controller: SelectionControllerYou},
			want:      "each creature you control",
		},
		{
			name:      "attacking creature",
			selection: SelectionSyntax{Kind: SelectionCreature, Attacking: true},
			want:      "each attacking creature",
		},
		{
			name:      "noncreature artifact",
			selection: SelectionSyntax{Kind: SelectionArtifact, ExcludedTypes: []CardType{CardTypeCreature}},
			want:      "each noncreature artifact",
		},
		{
			// A numeric rider is reconstructed verbatim and is exempt from the
			// selectionPhrase cross-check, which orders the numeric qualifier
			// ahead of the controller clause.
			name:      "creature you control with power numeric",
			selection: SelectionSyntax{Kind: SelectionCreature, Controller: SelectionControllerYou, MatchPower: true, Power: compare.Int{Op: compare.GreaterOrEqual, Value: 4}},
			want:      "each creature you control with power 4 or greater",
		},
		{
			// A subtype recipient is owned by the bespoke reconstruction;
			// selectionPhrase fails closed for it and the cross-check defers.
			name:      "subtype goblin",
			selection: SelectionSyntax{SubtypesAny: []types.Sub{types.Sub("Goblin")}},
			want:      "each Goblin",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := exactGroupDamagePermanentRecipientText(tc.selection)
			if !ok {
				t.Fatalf("exactGroupDamagePermanentRecipientText(%s) ok = false, want true", tc.name)
			}
			if got != tc.want {
				t.Errorf("exactGroupDamagePermanentRecipientText(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

// TestSelectionPhraseVerifiesGroupRecipientGap proves the soft gate closes the
// drift gap for the modeled forms while deferring to the bespoke reconstruction
// for the forms the renderer cannot represent. A modeled recipient that
// disagrees with the typed selection is rejected; a numeric or counter recipient
// is exempted; and an unrepresentable selection defers (the gate passes so the
// bespoke text stands alone).
func TestSelectionPhraseVerifiesGroupRecipientGap(t *testing.T) {
	t.Parallel()
	you := SelectionSyntax{Kind: SelectionCreature, Controller: SelectionControllerYou}

	if !selectionPhraseVerifiesGroupRecipient(you, "each creature you control", "", nil) {
		t.Error("matching recipient should pass the gate")
	}
	if selectionPhraseVerifiesGroupRecipient(you, "each creature", "", nil) {
		t.Error("recipient dropping the controller clause should fail the gate")
	}

	// A trailing numeric rider exempts the recipient from the cross-check.
	if !selectionPhraseVerifiesGroupRecipient(you, "each creature you control", " with power 4 or greater", nil) {
		t.Error("numeric-rider recipient should be exempt from the gate")
	}
	// A counter clause exempts the recipient from the cross-check.
	if !selectionPhraseVerifiesGroupRecipient(you, "each creature you control", "", []string{"with", "a", "counter", "on", "it"}) {
		t.Error("counter-clause recipient should be exempt from the gate")
	}

	// An unrepresentable selection (keyword) defers: selectionPhrase fails
	// closed, so the gate passes and the bespoke reconstruction stands alone.
	keyword := SelectionSyntax{Kind: SelectionCreature, Keyword: KeywordFlying}
	if !selectionPhraseVerifiesGroupRecipient(keyword, "each creature with flying", "", nil) {
		t.Error("unrepresentable selection should defer to the bespoke reconstruction")
	}
}
