package parser

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// foldTrailingImpulseExile folds a top-of-library exile that trails a resolving
// sentence together with the play/cast permission sentence that immediately
// follows it into a single EffectImpulseExile, leaving any effects that precede
// the exile in place.
//
// It is the generalized sibling of recognizeImpulseExileSequence, which folds
// only a body that is exactly the two-sentence "Exile ... library. Until end of
// turn, you may play/cast ..." shape. This one handles the exile appearing as
// the trailing conjunct of a compound sentence ("create a Treasure token and
// exile the top card of that player's library. Until end of turn, you may cast
// that card.", Ragavan, Nimble Pilferer), or as a standalone sentence preceded
// by other resolving sentences. The folded impulse becomes the last effect of
// its sentence and the preceding effects sequence ahead of it, so the ordered
// effect lowerer emits them in order.
//
// It credits nothing and returns ok=false unless the exile is the last effect of
// some sentence, that exile's clause is exactly a recognized top-of-library
// exile, and the very next sentence is exactly the matching single-effect
// play/cast permission. On success it returns the number of folded legacy and
// current effects (the consumed permission sentence) so the caller decrements
// its running totals, matching the credit-rider fold contract.
func foldTrailingImpulseExile(sentences []Sentence, atoms Atoms) (foldedLegacy, foldedEffects int, ok bool) {
	for p := 1; p < len(sentences); p++ {
		prev := &sentences[p-1]
		perm := &sentences[p]
		if len(prev.Effects) == 0 || len(perm.Effects) != 1 {
			continue
		}
		exile := &prev.Effects[len(prev.Effects)-1]
		if !impulseExileFoldExileCandidate(exile) {
			continue
		}
		// recognizeImpulseExileSequence only folds a body whose exile sentence is
		// exactly "Exile the top card of your library." and so never carries a
		// condition. This generalized fold matches only the clean exile clause
		// text, with any leading OR trailing condition bounded off into a separate
		// condition clause; folding such a sentence collapses it to the pure
		// two-sentence impulse shape that stripImpulseExileSemantics clears of
		// every condition clause, silently dropping the gate (e.g. Impossible
		// Inferno's Delirium condition, or a trailing "... library if <predicate>").
		// Fail the fold closed on any condition so such a card is left unsupported
		// rather than generated ungated.
		if sentenceHasConditionClause(prev) {
			continue
		}
		exileText := strings.TrimSpace(exactEffectClauseText(exile))
		clause, cok := matchImpulseExileClause(exileText)
		if !cok {
			continue
		}
		objectAmount := clause.amount
		if clause.variableX {
			objectAmount = 2
		}
		permText := strings.TrimSpace(perm.Text)
		permission, pok := matchImpulsePlayPermissionClause(permText, objectAmount)
		if !pok {
			continue
		}
		span := shared.Span{Start: exile.ClauseSpan.Start, End: perm.Span.End}
		tokens := append(append([]shared.Token(nil), exile.Tokens...), perm.Tokens...)
		*exile = EffectSyntax{
			Kind:                 EffectImpulseExile,
			Context:              clause.owner,
			Span:                 span,
			ClauseSpan:           span,
			Text:                 exileText + " " + permText,
			Tokens:               tokens,
			Amount:               EffectAmountSyntax{Value: clause.amount, Known: !clause.variableX, VariableX: clause.variableX},
			Duration:             permission.duration,
			ImpulseCast:          permission.cast,
			ImpulseSpendAnyColor: permission.spendAnyColor,
			Exact:                true,
		}
		foldedEffects = len(perm.Effects)
		if perm.LegacyEffects {
			foldedLegacy = orderedEffectCount(semanticEffectTokens(perm.Tokens), atoms)
		}
		perm.Effects = nil
		perm.LegacyEffects = false
		perm.ImpulseExilePermission = true
		return foldedLegacy, foldedEffects, true
	}
	return 0, 0, false
}

// impulseExileFoldExileCandidate reports whether effect is a clean top-of-library
// exile eligible to fold with a following play/cast permission: a mandatory,
// non-negated, non-additional exile with no targets of its own. The clause text
// is matched separately by matchImpulseExileClause; this guards the structural
// shape so a modified or targeted exile never folds into an impulse.
func impulseExileFoldExileCandidate(effect *EffectSyntax) bool {
	return effect.Kind == EffectExile &&
		!effect.Optional &&
		!effect.Negated &&
		!effect.Additional &&
		len(effect.Targets) == 0
}

// sentenceHasConditionClause reports whether the sentence carries a condition
// clause introducer ("If ...", "Unless ...", "As long as ...", "While ...",
// "During your turn,", reflexive "When you do,") at any position — leading or
// trailing. A trailing condition ("Exile the top card of your library if
// <predicate>.") is bounded off the effect clause the fold matches, so
// exactEffectClauseText returns the clean exile text and the fold would
// otherwise fire and let stripImpulseExileSemantics drop the recorded gate. The
// fold uses this to fail closed on any gated exile sentence, leaving such a card
// unsupported rather than generating it ungated.
func sentenceHasConditionClause(sentence *Sentence) bool {
	tokens := semanticEffectTokens(sentence.Tokens)
	for i := range tokens {
		if kind, _ := conditionIntroAt(tokens, i); kind != ConditionIntroUnknown {
			return true
		}
	}
	return false
}
