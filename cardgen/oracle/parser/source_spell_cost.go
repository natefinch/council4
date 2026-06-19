package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// emitSourceSpellCostReduction marks the EffectCast effect of the exact
// single-clause ability "This spell costs {N} less to cast for each <countable
// battlefield object>." as a typed source-scoped cast cost reduction. The
// resolving-syntax pass already classifies the clause as an EffectCast whose
// typed Amount holds the per-object battlefield count; this pass confirms the
// exact wording, captures the per-object generic reduction N from the {N}
// symbol, and records it on the effect so lowering can build a source-scoped
// cost modifier without re-reading source text. Any wording that does not match
// the exact shape, or whose counted objects are not battlefield permanents, is
// left untouched so it stays unsupported and fails closed.
func emitSourceSpellCostReduction(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.Modal != nil {
			continue
		}
		if len(ability.Sentences) != 1 || len(ability.Sentences[0].Effects) != 1 {
			continue
		}
		effect := &ability.Sentences[0].Effects[0]
		if effect.Kind != EffectCast || effect.Context != EffectContextSource {
			continue
		}
		if effect.Amount.DynamicKind != EffectDynamicAmountCount ||
			effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
			effect.Amount.Multiplier != 1 ||
			effect.Amount.Selection == nil {
			continue
		}
		tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
		amount, ok := sourceSpellCostReductionAmount(tokens, ability.Atoms)
		if !ok {
			continue
		}
		effect.SourceSpellCostReduction = true
		effect.SourceSpellCostReductionAmount = amount
	}
}

// sourceSpellCostReductionAmount validates the exact "This spell costs {N} less
// to cast for each <count subject>." wording and returns the per-object generic
// reduction N. The subject phrase must be the spell itself ("This spell" or the
// card's own name) and the counted objects must be battlefield permanents the
// existing typed count machinery represents; graveyard, hand, variable {X}, and
// any other shape fail closed by returning false.
func sourceSpellCostReductionAmount(tokens []shared.Token, atoms Atoms) (int, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return 0, false
	}
	idx, ok := sourceSpellSubjectEnd(tokens, atoms)
	if !ok {
		return 0, false
	}
	if !effectWordsAt(tokens, idx, "costs") {
		return 0, false
	}
	idx++
	if idx >= len(tokens) || tokens[idx].Kind != shared.Symbol {
		return 0, false
	}
	amount, ok := staticGenericSymbolValue(tokens[idx].Text)
	if !ok || amount <= 0 {
		return 0, false
	}
	idx++
	if !effectWordsAt(tokens, idx, "less", "to", "cast", "for", "each") {
		return 0, false
	}
	idx += 5
	if idx >= len(tokens)-1 {
		return 0, false
	}
	subject, ok := parseDynamicCountSubject(tokens, idx, atoms)
	if !ok || !subject.count {
		return 0, false
	}
	if subject.end != len(tokens)-1 {
		return 0, false
	}
	if subject.amount.DynamicKind != EffectDynamicAmountCount || subject.amount.Selection == nil {
		return 0, false
	}
	if subject.amount.Selection.Zone != zone.None {
		return 0, false
	}
	return amount, true
}

// sourceSpellSubjectEnd returns the token index just past a leading source-spell
// subject ("This spell" or the card's own name), and whether one was present.
func sourceSpellSubjectEnd(tokens []shared.Token, atoms Atoms) (int, bool) {
	if effectWordsAt(tokens, 0, "this", "spell") {
		return 2, true
	}
	nameSpan, ok := atoms.SelfNameSpanStartingAt(tokens[0].Span)
	if !ok {
		return 0, false
	}
	end := 0
	for end < len(tokens) && tokens[end].Span.End.Offset <= nameSpan.End.Offset {
		end++
	}
	if end == 0 {
		return 0, false
	}
	return end, true
}
