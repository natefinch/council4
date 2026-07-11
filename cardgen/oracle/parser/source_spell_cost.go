package parser

import (
	"strings"

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
		effect := singleResolvingEffect(ability)
		if effect == nil {
			continue
		}
		if effect.Kind != EffectCast || effect.Context != EffectContextSource {
			continue
		}
		party := effect.Amount.DynamicKind == EffectDynamicAmountPartySize
		if effect.Amount.DynamicKind != EffectDynamicAmountCount && !party ||
			effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
			effect.Amount.Multiplier != 1 ||
			!party && effect.Amount.Selection == nil {
			continue
		}
		amount, ok := sourceSpellCostReductionAmount(sourceSpellCostBodyTokens(ability), ability.Atoms)
		if !ok {
			continue
		}
		if party {
			effect.SourceSpellCostReductionDynamic = true
		} else {
			effect.SourceSpellCostReduction = true
		}
		effect.SourceSpellCostReductionAmount = amount
	}
}

// emitSourceSpellCostReductionDynamic marks the EffectCast effect of the exact
// single-clause ability "This spell costs {X} less to cast, where X is <dynamic
// amount>." as a typed source-scoped cast cost reduction whose amount is the
// effect's own typed dynamic Amount (The Great Henge: the greatest power among
// creatures you control). The resolving-syntax pass already captures the
// "where X is ..." dynamic on the effect's Amount; this pass confirms the exact
// "costs {X} less to cast," framing and that the captured dynamic is a kind
// lowering can evaluate at cost time, then records the marker so lowering can
// build a source-scoped cost modifier without re-reading source text. Any other
// wording or dynamic shape is left untouched so it stays unsupported and fails
// closed.
func emitSourceSpellCostReductionDynamic(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.Modal != nil {
			continue
		}
		effect := singleResolvingEffect(ability)
		if effect == nil {
			continue
		}
		if effect.Kind != EffectCast || effect.Context != EffectContextSource {
			continue
		}
		if effect.Amount.DynamicForm != EffectDynamicAmountFormWhereX ||
			!sourceSpellCostReductionDynamicKind(effect.Amount.DynamicKind) {
			continue
		}
		if !sourceSpellCostReductionDynamicFrame(sourceSpellCostBodyTokens(ability), ability.Atoms) {
			continue
		}
		effect.SourceSpellCostReductionDynamic = true
	}
}

// emitSourceSpellCostReductionConditional marks the EffectCast effect of the
// exact single-clause ability "This spell costs {N} less to cast if
// <condition>." as a typed source-scoped flat cast cost reduction gated by the
// ability's condition clause (Wizard's Lightning: "if you control a Wizard";
// Squash: "if you control a Giant"; Draconic Lore: "if you control a Dragon").
// The resolving-syntax pass classifies the clause as a source-context EffectCast
// with no typed Amount; the condition-clause pass already captured the trailing
// "if ..." predicate on the ability. This pass confirms the exact "This spell
// costs {N} less to cast" framing, captures the flat generic reduction N, and
// records the marker so lowering can build a source-scoped conditional cost
// modifier without re-reading source text. Any wording that does not match the
// exact framing, that carries a dynamic Amount (the per-object or "where X is"
// forms own that wording), or that has no recognized condition clause is left
// untouched so it stays unsupported and fails closed.
func emitSourceSpellCostReductionConditional(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		targetsTappedCreature := strings.HasSuffix(
			strings.ToLower(strings.TrimSpace(ability.Text)),
			"if it targets a tapped creature.",
		)
		if ability.Modal != nil || len(ability.ConditionClauses) == 0 && !targetsTappedCreature {
			continue
		}
		effect := singleResolvingEffect(ability)
		if effect == nil {
			continue
		}
		if effect.Kind != EffectCast || effect.Context != EffectContextSource {
			continue
		}
		if effect.SourceSpellCostReduction || effect.SourceSpellCostReductionDynamic {
			continue
		}
		if effect.Amount.DynamicKind != "" || effect.Amount.DynamicForm != "" {
			continue
		}
		amount, ok := sourceSpellCostReductionFlatAmount(sourceSpellCostBodyTokens(ability), ability.Atoms)
		if !ok {
			continue
		}
		effect.SourceSpellCostReductionConditional = true
		effect.SourceSpellCostReductionAmount = amount
		effect.SourceSpellCostReductionTargetsTappedCreature = targetsTappedCreature
	}
}

// sourceSpellCostReductionFlatAmount validates the exact "This spell costs {N}
// less to cast" framing of a conditional source-spell cost reduction and returns
// the flat generic reduction N. The subject must be the spell itself ("This
// spell" or the card's own name), the cost symbol must be a fixed generic {N},
// and a condition introducer ("if", "as long as", "only if", "unless") must
// follow the "cast" verb so the trailing predicate is owned by the ability's
// condition clause. Any other shape fails closed by returning false.
func sourceSpellCostReductionFlatAmount(tokens []shared.Token, atoms Atoms) (int, bool) {
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
	if !effectWordsAt(tokens, idx, "less", "to", "cast") {
		return 0, false
	}
	idx += 3
	if idx >= len(tokens) {
		return 0, false
	}
	intro, _ := conditionIntroAt(tokens, idx)
	return amount, intro != ConditionIntroUnknown
}

// singleResolvingEffect returns the ability's only resolving effect when exactly
// one sentence carries effects, that sentence carries exactly one effect, and
// every other sentence is reminder-only (no effects, static rule, targets, or
// payment prelude). A reminder such as the "(Artifacts, legendaries, and Sagas
// are historic.)" that follows The Capitoline Triad's cost reduction parses as
// its own effect-free sentence; this tolerates it while a genuine second
// resolving clause still fails closed by returning nil.
func singleResolvingEffect(ability *Ability) *EffectSyntax {
	var found *EffectSyntax
	for s := range ability.Sentences {
		sentence := &ability.Sentences[s]
		switch len(sentence.Effects) {
		case 0:
			if sentence.StaticRule != nil || len(sentence.Targets) != 0 || sentence.PaymentPrelude != nil {
				return nil
			}
		case 1:
			if found != nil {
				return nil
			}
			found = &sentence.Effects[0]
		default:
			return nil
		}
	}
	return found
}

// sourceSpellCostBodyTokens returns the ability's resolving-body semantic tokens:
// the tokens within BodySpan (after any ability-word or chapter prefix) with
// reminder and quoted spans removed. The Capitoline Triad's "Those Who Came
// Before —" ability-word prefix and historic reminder are excluded so the
// subject scan begins at "This spell".
func sourceSpellCostBodyTokens(ability *Ability) []shared.Token {
	body := tokensWithinParserSpan(ability.Tokens, ability.BodySpan)
	return eventHistorySemanticTokens(body, ability.Reminders, ability.Quoted)
}

// sourceSpellCostReductionDynamicKind reports whether a "where X is ..." dynamic
// amount kind can scale a source-spell cost reduction: the controller-aggregate
// and battlefield-group kinds the runtime evaluates at cost time without a
// resolving stack object. Object-referencing kinds fail closed.
func sourceSpellCostReductionDynamicKind(kind EffectDynamicAmountKind) bool {
	switch kind {
	case EffectDynamicAmountCount,
		EffectDynamicAmountGreatestPower,
		EffectDynamicAmountGreatestToughness,
		EffectDynamicAmountGreatestManaValue,
		EffectDynamicAmountTotalPower,
		EffectDynamicAmountTotalToughness,
		EffectDynamicAmountTotalManaValue,
		EffectDynamicAmountControllerLife,
		EffectDynamicAmountOpponentCount,
		EffectDynamicAmountBasicLandTypes,
		EffectDynamicAmountDevotion:
		return true
	default:
		return false
	}
}

// sourceSpellCostReductionDynamicFrame validates the exact "This spell costs {X}
// less to cast, where X is ..." framing. The subject must be the spell itself
// ("This spell" or the card's own name), the cost symbol must be the variable
// {X}, and the dynamic clause must open with a comma followed by "where". The
// trailing dynamic amount is validated by the typed Amount the caller already
// checked.
func sourceSpellCostReductionDynamicFrame(tokens []shared.Token, atoms Atoms) bool {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return false
	}
	idx, ok := sourceSpellSubjectEnd(tokens, atoms)
	if !ok {
		return false
	}
	if !effectWordsAt(tokens, idx, "costs") {
		return false
	}
	idx++
	if idx >= len(tokens) || tokens[idx].Kind != shared.Symbol {
		return false
	}
	if symbol, ok := staticTrimSymbol(tokens[idx].Text); !ok || symbol != "X" {
		return false
	}
	idx++
	if !effectWordsAt(tokens, idx, "less", "to", "cast") {
		return false
	}
	idx += 3
	if idx >= len(tokens) || tokens[idx].Kind != shared.Comma {
		return false
	}
	idx++
	return effectWordsAt(tokens, idx, "where")
}

// sourceSpellCostReductionAmount validates the exact "This spell costs {N} less
// to cast for each <count subject>." wording and returns the per-object generic
// reduction N. The subject phrase must be the spell itself ("This spell" or the
// card's own name) and the counted objects must be battlefield permanents, or
// cards in the caster's own graveyard or hand, that the typed count machinery
// represents; library, exile, variable {X}, and any other shape fail closed by
// returning false.
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
	if subject.amount.DynamicKind == EffectDynamicAmountPartySize {
		return amount, true
	}
	if subject.amount.DynamicKind != EffectDynamicAmountCount || subject.amount.Selection == nil {
		return 0, false
	}
	switch subject.amount.Selection.Zone {
	case zone.None, zone.Graveyard, zone.Hand:
	default:
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
