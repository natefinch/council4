package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// odricKeywordShareText is Odric, Lunarch Marshal's exact oracle text.
const odricKeywordShareText = "At the beginning of each combat, creatures you control gain first strike until end of turn if a creature you control has first strike. The same is true for flying, deathtouch, double strike, haste, hexproof, indestructible, lifelink, menace, reach, skulk, trample, and vigilance."

// firstKeywordShare returns the first compiled KeywordShare across a
// compilation's abilities, or nil when none was threaded through.
func firstKeywordShare(compilation Compilation) *CompiledKeywordShare {
	for i := range compilation.Abilities {
		if compilation.Abilities[i].KeywordShare != nil {
			return compilation.Abilities[i].KeywordShare
		}
	}
	return nil
}

// TestCompileKeywordShareThreadsOdric proves the compiler carries the recognized
// team keyword-sharing construct through as a CompiledKeywordShare holding the
// ordered keyword kinds, with no unsupported diagnostic for the folded body.
func TestCompileKeywordShareThreadsOdric(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(odricKeywordShareText, pipelineContext{CardName: "Odric, Lunarch Marshal"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	share := firstKeywordShare(compilation)
	if share == nil {
		t.Fatalf("compiled abilities carried no KeywordShare; abilities = %#v", compilation.Abilities)
	}
	want := []parser.KeywordKind{
		parser.KeywordFirstStrike,
		parser.KeywordFlying,
		parser.KeywordDeathtouch,
		parser.KeywordDoubleStrike,
		parser.KeywordHaste,
		parser.KeywordHexproof,
		parser.KeywordIndestructible,
		parser.KeywordLifelink,
		parser.KeywordMenace,
		parser.KeywordReach,
		parser.KeywordSkulk,
		parser.KeywordTrample,
		parser.KeywordVigilance,
	}
	if len(share.Keywords) != len(want) {
		t.Fatalf("keywords = %#v, want %d", share.Keywords, len(want))
	}
	for i, kind := range want {
		if share.Keywords[i] != kind {
			t.Fatalf("keyword %d = %q, want %q", i, share.Keywords[i], kind)
		}
	}
}

// TestCompileKeywordShareAbsentForSiblings proves the compiler threads no
// KeywordShare for cards the parser leaves unrecognized, so the fail-closed
// siblings never present as the typed construct.
func TestCompileKeywordShareAbsentForSiblings(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		source string
	}{
		{
			name:   "Concerted Effort",
			source: "At the beginning of each upkeep, creatures you control gain flying until end of turn if a creature you control has flying. The same is true for fear, first strike, double strike, landwalk, protection, trample, and vigilance.",
		},
		{
			name:   "Bleeding Effect",
			source: "At the beginning of combat on your turn, creatures you control gain flying until end of turn if a creature card in your graveyard has flying. The same is true for first strike, double strike, deathtouch, hexproof, indestructible, lifelink, menace, reach, trample, and vigilance.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			compilation, _ := compileSource(tc.source, pipelineContext{CardName: tc.name})
			if share := firstKeywordShare(compilation); share != nil {
				t.Fatalf("%s threaded a KeywordShare; keywords = %#v", tc.name, share.Keywords)
			}
		})
	}
}
