package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestParseIdentityTokenCreationReplacement(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"If one or more creature tokens would be created under your control, that many 4/4 white Angel creature tokens with flying and vigilance are created instead.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if len(ability.SemanticKeywords) != 0 {
		t.Fatalf("semantic keywords = %#v, want none (substitute keywords belong to the token)", ability.SemanticKeywords)
	}
	effects := ability.Sentences[0].Effects
	if len(effects) != 2 {
		t.Fatalf("effects = %#v, want two (would-create group + identity output)", effects)
	}
	if effects[0].Replacement.Kind != EffectReplacementNone {
		t.Fatalf("would-create group replacement kind = %v, want none", effects[0].Replacement.Kind)
	}
	output := effects[1]
	if output.Replacement.Kind != EffectReplacementThatManyIdentity {
		t.Fatalf("output replacement kind = %v, want EffectReplacementThatManyIdentity", output.Replacement.Kind)
	}
	if !output.TokenPTKnown || output.TokenPower != 4 || output.TokenToughness != 4 {
		t.Fatalf("output P/T = %d/%d (known %v), want 4/4 known", output.TokenPower, output.TokenToughness, output.TokenPTKnown)
	}
	if got := output.Selection.SubtypesAny; !slices.Equal(got, []types.Sub{types.Angel}) {
		t.Fatalf("output subtypes = %v, want [Angel]", got)
	}
	wantKeywords := []KeywordKind{KeywordFlying, KeywordVigilance}
	if !slices.Equal(output.TokenKeywords, wantKeywords) {
		t.Fatalf("output token keywords = %v, want %v", output.TokenKeywords, wantKeywords)
	}
}

func TestParseIdentityTokenCreationReplacementRejectsDoubling(t *testing.T) {
	t.Parallel()
	commaIndex, anyController, ok := matchPassiveTokenIdentity(
		tokensOf(t, "If one or more tokens would be created under your control, twice that many of those tokens are created instead."),
	)
	if ok {
		t.Fatalf("matchPassiveTokenIdentity matched doubling wording: commaIndex=%d anyController=%v", commaIndex, anyController)
	}
}

func tokensOf(t *testing.T, source string) []shared.Token {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	return document.Abilities[0].Sentences[0].Tokens
}
