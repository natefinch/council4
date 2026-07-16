package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// recognizeConditionalImpulseExileSequence folds a top-card exile followed by
// conditional free-play and otherwise normal-play permissions into two atomic
// impulse-exile effects. Ordinary sequence gating then runs exactly one branch,
// so the card is exiled only once.
func recognizeConditionalImpulseExileSequence(sentences []Sentence) bool {
	if len(sentences) != 3 ||
		len(sentences[0].Effects) != 1 ||
		len(sentences[1].Effects) != 1 ||
		len(sentences[2].Effects) != 0 {
		return false
	}
	exile := sentences[0].Effects[0]
	freePlay := sentences[1].Effects[0]
	if !impulseExileFoldExileCandidate(&exile) ||
		!exile.Exact ||
		!sentenceHasConditionClause(&sentences[1]) ||
		freePlay.Kind != EffectPlay ||
		!freePlay.Exact ||
		!freePlay.Optional ||
		freePlay.Negated ||
		freePlay.Context != EffectContextController ||
		!freePlay.CastWithoutPayingManaCost ||
		freePlay.Duration != EffectDurationThisTurn ||
		len(freePlay.References) != 1 ||
		freePlay.References[0].Kind != ReferenceThatObject {
		return false
	}
	exileClause, ok := matchImpulseExileClause(strings.TrimSpace(exactEffectClauseText(&exile)))
	if !ok || exileClause.variableX || exileClause.amount != 1 {
		return false
	}
	otherwiseText := strings.TrimSpace(sentences[2].Text)
	otherwiseText = strings.TrimSpace(strings.TrimPrefix(otherwiseText, "Otherwise,"))
	otherwise, ok := matchImpulsePlayPermissionClause(otherwiseText, 1)
	if !ok ||
		otherwise.cast ||
		otherwise.spendAnyColor ||
		otherwise.duration != EffectDurationThisTurn {
		return false
	}

	freeSpan := shared.Span{Start: exile.ClauseSpan.Start, End: sentences[1].Span.End}
	freeTokens := append(append([]shared.Token(nil), exile.Tokens...), sentences[1].Tokens...)
	freeImpulse := EffectSyntax{
		Kind:                         EffectImpulseExile,
		Context:                      exileClause.owner,
		Span:                         freeSpan,
		ClauseSpan:                   freeSpan,
		Text:                         strings.TrimSpace(sentences[0].Text) + " " + strings.TrimSpace(sentences[1].Text),
		Tokens:                       freeTokens,
		Amount:                       EffectAmountSyntax{Value: 1, Known: true},
		Duration:                     EffectDurationThisTurn,
		ImpulseWithoutPayingManaCost: true,
		Exact:                        true,
	}
	otherwiseImpulse := EffectSyntax{
		Kind:       EffectImpulseExile,
		Context:    exileClause.owner,
		Connection: EffectConnectionOtherwise,
		Span:       sentences[2].Span,
		ClauseSpan: sentences[2].Span,
		Text:       sentences[2].Text,
		Tokens:     cloneTokens(sentences[2].Tokens),
		Amount:     EffectAmountSyntax{Value: 1, Known: true},
		Duration:   EffectDurationThisTurn,
		Exact:      true,
	}
	sentences[0].Effects = nil
	sentences[0].LegacyEffects = false
	sentences[1].Effects = []EffectSyntax{freeImpulse}
	sentences[2].Effects = []EffectSyntax{otherwiseImpulse}
	return true
}
