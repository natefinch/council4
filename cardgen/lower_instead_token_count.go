package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerInsteadOfThoseTokensSequence lowers the "create A tokens. If <condition>,
// create N of those tokens instead." kicker/replacement family (Rite of
// Replication, Saproling Migration, Conqueror's Pledge, Increasing Devotion) to
// two mutually exclusive CreateToken instructions that share the first clause's
// token source.
//
// The first clause fully describes the token (a copy of a target, or a
// synthesized typed token); the second clause names the same tokens only as
// "those tokens" and supplies a new count. The reflexive backreference carries
// no token characteristics of its own, so the same token source is reused for
// both counts. The "instead" replacement makes the two counts mutually
// exclusive: the base count resolves when the gating condition is false and the
// replacement count resolves when it is true. It fails closed for every other
// shape (a second clause with its own token characteristics, a missing or
// non-gating condition, a dynamic replacement count, or an unsupported first
// clause).
func lowerInsteadOfThoseTokensSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 1 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	base := ctx.content.Effects[0]
	replacement := ctx.content.Effects[1]
	if base.Kind != compiler.EffectCreate ||
		replacement.Kind != compiler.EffectCreate ||
		base.Context != parser.EffectContextController ||
		replacement.Context != parser.EffectContextController ||
		base.Replacement.Kind != parser.EffectReplacementNone ||
		replacement.Replacement.Kind != parser.EffectReplacementInstead {
		return game.AbilityContent{}, false
	}
	condition := ctx.content.Conditions[0]
	if !spanCovered(condition.Span, []shared.Span{replacement.Span}) ||
		spanCovered(condition.Span, []shared.Span{base.Span}) {
		return game.AbilityContent{}, false
	}
	if !bareOfThoseTokensCreate(&replacement, condition.Span) {
		return game.AbilityContent{}, false
	}

	gate, ok := effectGateCondition(condition)
	if !ok {
		return game.AbilityContent{}, false
	}
	negated, ok := negatedEffectCondition(&gate)
	if !ok {
		return game.AbilityContent{}, false
	}

	baseInstruction, mode, ok := lowerSingleCreateInstruction(ctx, base)
	if !ok {
		return game.AbilityContent{}, false
	}
	// The replacement count copies the first clause verbatim and substitutes only
	// the count, so its token source matches the base clause's exactly.
	replacementClause := base
	replacementClause.Amount = replacement.Amount
	replacementInstruction, _, ok := lowerSingleCreateInstruction(ctx, replacementClause)
	if !ok {
		return game.AbilityContent{}, false
	}

	baseInstruction.Condition = opt.Val(negated)
	replacementInstruction.Condition = opt.Val(gate)
	return game.Mode{
		Targets:  mode.Targets,
		Sequence: []game.Instruction{baseInstruction, replacementInstruction},
	}.Ability(), true
}

// bareOfThoseTokensCreate reports whether effect is a "create N of those tokens"
// backreference: a controller create whose count is a fixed positive integer and
// which carries no token-defining characteristics of its own. Its only reference
// outside the gating condition's span must be the plural "those" pronoun that
// names the tokens a prior clause created.
func bareOfThoseTokensCreate(effect *compiler.CompiledEffect, conditionSpan shared.Span) bool {
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicForm != compiler.DynamicAmountFormNone ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		return false
	}
	if effect.Selector.Kind != compiler.SelectorUnknown ||
		effect.TokenCopyOfTarget ||
		effect.TokenCopyOfReference ||
		effect.TokenCopyOfAttached ||
		effect.TokenCopyOfTriggeringSet ||
		effect.TokenPTKnown ||
		effect.TokenPTVariableX ||
		effect.TokenChoice ||
		effect.Negated ||
		effect.TokenName != "" ||
		effect.TokenPredefinedName != "" ||
		effect.AmassSubtype != "" ||
		len(effect.Targets) != 0 ||
		len(effect.TokenKeywords) != 0 ||
		effect.TokenGrantedAbility != nil {
		return false
	}
	owned := referencesOutsideSpan(effect.References, conditionSpan)
	if len(owned) != 1 ||
		owned[0].Kind != compiler.ReferencePronoun ||
		owned[0].Pronoun != compiler.ReferencePronounThose {
		return false
	}
	return true
}

// effectGateCondition lowers a single sequence condition to its effect-gate form,
// the wrapper the "instead" replacement gates the two token counts on.
func effectGateCondition(condition compiler.CompiledCondition) (game.EffectCondition, bool) {
	lowered, ok := lowerCondition(condition, conditionContextEffectGate)
	if !ok {
		return game.EffectCondition{}, false
	}
	return game.EffectCondition{Condition: opt.Val(lowered)}, true
}

// lowerSingleCreateInstruction lowers a single create effect through the shared
// token-creation entry point and returns its one CreateToken instruction along
// with the mode (which carries any target spec for a copy-of-target create). It
// fails closed unless the effect lowers to exactly one instruction.
func lowerSingleCreateInstruction(ctx contentCtx, effect compiler.CompiledEffect) (game.Instruction, game.Mode, bool) {
	clauseCtx := contextForEffect(ctx, &effect)
	clauseCtx.content.Conditions = nil
	content, diagnostic := lowerCreateTokenSpellLinked(clauseCtx, "")
	if diagnostic != nil ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Sequence) != 1 ||
		content.Modes[0].Sequence[0].Condition.Exists {
		return game.Instruction{}, game.Mode{}, false
	}
	return content.Modes[0].Sequence[0], content.Modes[0], true
}
