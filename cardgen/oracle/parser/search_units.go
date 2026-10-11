package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// Analyze a typed contiguous search group independently of later clauses in
// the same sentence. Every parameter still uses the existing exact grammar.
func recognizeSearchUnits(sentences []Sentence) {
	for si := range sentences {
		effects := sentences[si].Effects
		for start := range effects {
			if effects[start].Kind != EffectSearch {
				continue
			}
			put := start + 1
			if put < len(effects) && effects[put].Kind == EffectReveal {
				put++
			}
			shuffle := put + 1
			if shuffle+1 >= len(effects) || effects[put].Kind != EffectPut ||
				effects[shuffle].Kind != EffectShuffle || effects[shuffle].Connection != EffectConnectionThen ||
				effects[shuffle+1].Connection != EffectConnectionThen {
				continue
			}
			search := &effects[start]
			if len(search.Tokens) == 0 {
				continue
			}
			tokens := sentences[si].Tokens
			begin := slices.IndexFunc(tokens, func(token shared.Token) bool {
				return token.Span == search.Tokens[0].Span
			})
			end := slices.IndexFunc(tokens, func(token shared.Token) bool {
				return token.Span.Start.Offset >= effects[shuffle+1].ConnectionSpan.Start.Offset
			})
			if begin < 0 || end < begin {
				continue
			}
			for end > begin && tokens[end-1].Kind == shared.Comma {
				end--
			}
			clipped := *search
			clipped.Tokens = tokens[begin:end]
			clipped.Text = joinedEffectText(clipped.Tokens) + "."
			analysis := analyzeSearchClause(&clipped)
			if analysis.detail != "" || analysis.sharedSubtype || effects[put].SearchSplit.Present {
				continue
			}
			search.Exact = true
			search.UnsupportedDetail = ""
			search.SearchSharedSubtype = analysis.sharedSubtype
			search.SearchDifferentNames = analysis.differentNames
			search.SearchSameNameAsChosenObject = analysis.sameNameAsChosenObject
			search.SearchDestination = analysis.destinationPosition
			search.SearchControl = analysis.control
			search.SearchSlots = searchHeterogeneousSlotSubtypes(&clipped)
		}
	}
}
