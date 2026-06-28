package parser

import "testing"

// TestParseMillHalfLibraryRoundsDownAndUp verifies the "mills half their
// library, rounded up/down" family (Traumatize, Fleet Swallower) types to the
// half-player-library dynamic amount with the rounding direction recorded, a
// card selection, and an exact round-trip, for both the target-player and
// defending-player subjects.
func TestParseMillHalfLibraryRoundsDownAndUp(t *testing.T) {
	t.Parallel()
	cases := []struct {
		text        string
		instSorcery bool
		wantRoundUp bool
		wantContext EffectContextKind
	}{
		{
			text:        "Target player mills half their library, rounded down.",
			instSorcery: true,
			wantRoundUp: false,
			wantContext: EffectContextTarget,
		},
		{
			text:        "Target opponent mills half their library, rounded up.",
			instSorcery: true,
			wantRoundUp: true,
			wantContext: EffectContextTarget,
		},
		{
			text:        "Whenever this creature attacks, defending player mills half their library, rounded up.",
			instSorcery: false,
			wantRoundUp: true,
			wantContext: EffectContextDefendingPlayer,
		},
	}
	for _, tc := range cases {
		document, diagnostics := Parse(tc.text, Context{InstantOrSorcery: tc.instSorcery})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", tc.text, diagnostics)
		}
		var mill *EffectSyntax
		for ai := range document.Abilities {
			for si := range document.Abilities[ai].Sentences {
				for ei := range document.Abilities[ai].Sentences[si].Effects {
					effect := &document.Abilities[ai].Sentences[si].Effects[ei]
					if effect.Kind == EffectMill {
						mill = effect
					}
				}
			}
		}
		if mill == nil {
			t.Fatalf("Parse(%q): no mill effect", tc.text)
		}
		if !mill.Exact {
			t.Fatalf("Parse(%q): mill not exact", tc.text)
		}
		if mill.Amount.DynamicKind != EffectDynamicAmountHalfPlayerLibrary ||
			mill.Amount.DynamicForm != EffectDynamicAmountFormHalfLibrary {
			t.Fatalf("Parse(%q): amount = %#v", tc.text, mill.Amount)
		}
		if mill.Amount.RoundUp != tc.wantRoundUp {
			t.Fatalf("Parse(%q): roundUp = %v, want %v", tc.text, mill.Amount.RoundUp, tc.wantRoundUp)
		}
		if mill.Selection.Kind != SelectionCard {
			t.Fatalf("Parse(%q): selection kind = %q, want SelectionCard", tc.text, mill.Selection.Kind)
		}
		if mill.Context != tc.wantContext {
			t.Fatalf("Parse(%q): context = %v, want %v", tc.text, mill.Context, tc.wantContext)
		}
	}
}

// TestParseMillHalfLibraryNotRecognizedWithoutRounding verifies the recognizer
// fails closed when the "rounded up/down" qualifier is absent, leaving the
// amount untyped so the clause stays unsupported rather than guessing a half.
func TestParseMillHalfLibraryNotRecognizedWithoutRounding(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Target player mills half their library.",
		Context{InstantOrSorcery: true},
	)
	for ai := range document.Abilities {
		for si := range document.Abilities[ai].Sentences {
			for _, effect := range document.Abilities[ai].Sentences[si].Effects {
				if effect.Kind == EffectMill &&
					effect.Amount.DynamicKind == EffectDynamicAmountHalfPlayerLibrary {
					t.Fatalf("unexpectedly recognized half-library without rounding: %#v", effect.Amount)
				}
			}
		}
	}
}
