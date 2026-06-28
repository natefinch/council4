package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// tokenCopyGrantRiderAttribution returns the keyword and reference compiler
// nodes that belong to a copy-token effect's folded "[That token] gains
// <keyword>." rider sentence. The parser folds that rider into the create
// effect's TokenCopyGrantKeywords, but its keyword and "that token" pronoun fall
// in a separate sentence whose span lies outside the create clause's own span.
// In an ordered effect sequence the per-effect keyword and reference
// attribution keys off the clause span, so without this the granted keyword and
// pronoun go unattributed: the copy-token lowerer rejects the keyword-count
// mismatch and the sequence's consumed-count check treats them as dropped. This
// is what blocks the temporary-copy family ("Create a token that's a copy of
// <target>. That token gains haste. Exile it at the beginning of the next end
// step.", Cogwork Assembler and kin), whose later cleanup clause already lowers
// on its own. It returns nothing for any effect that carries no grant rider, and
// excludes anything the create clause's own span already covers so a rider
// folded into the same sentence is never double-counted.
func tokenCopyGrantRiderAttribution(
	effect *compiler.CompiledEffect,
	keywords []compiler.CompiledKeyword,
	references []compiler.CompiledReference,
) ([]compiler.CompiledKeyword, []compiler.CompiledReference) {
	if len(effect.TokenCopyGrantKeywords) == 0 {
		return nil, nil
	}
	riderSpan := effect.TokenCopyGrantRiderSpan
	if riderSpan == (shared.Span{}) {
		return nil, nil
	}
	clauseSpans := []shared.Span{effect.ClauseSpan}
	var riderKeywords []compiler.CompiledKeyword
	for _, keyword := range keywordsWithinSpan(keywords, riderSpan) {
		if !spanCovered(keyword.Span, clauseSpans) {
			riderKeywords = append(riderKeywords, keyword)
		}
	}
	var riderReferences []compiler.CompiledReference
	for _, reference := range referencesWithinSpan(references, riderSpan) {
		if !spanCovered(reference.Span, clauseSpans) {
			riderReferences = append(riderReferences, reference)
		}
	}
	return riderKeywords, riderReferences
}
