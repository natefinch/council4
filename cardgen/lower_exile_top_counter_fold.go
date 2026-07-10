package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// foldExileTopNamedCounterSequence collapses the adjacent pair "exile ... from
// the top of your library ... Put a <name> counter on each of them." into a
// single top-of-library exile that carries the named exile counter, so the
// ordered-sequence lowering sees one exile effect (which the ExileTopOfLibrary
// primitive already realizes, placing the counter on every exiled card) instead
// of an unsupported standalone counter placement on cards resting in exile.
//
// The named counter is opaque here — this fold never inspects which counter it
// is — so any card that exiles library cards and puts its card-defined counter
// on each of them benefits (Flamewar's intel counters, and any future analog).
// It is restricted to the exact shape it can prove equivalent: the exile is a
// controller-scoped top-of-library source that does not already carry a counter,
// and the follow-up is an exact, non-optional, controller-context placement of a
// single known counter on the plural "them" back-reference naming the just-exiled
// batch, with no targets, duration, or conditions of its own. Every other shape
// falls through unchanged so the sequence lowering fails closed as before.
//
// It returns the (possibly rewritten) context and whether a fold occurred; the
// caller keeps the sequence path so the remaining effects (the trailing
// "Convert" transform) lower through the existing machinery.
func foldExileTopNamedCounterSequence(ctx contentCtx) (contentCtx, bool) {
	effects := ctx.content.Effects
	for i := 0; i+1 < len(effects); i++ {
		exile := effects[i]
		put := effects[i+1]
		if !exileTopNamedCounterFoldable(exile, put) {
			continue
		}
		folded := slices.Clone(effects)
		folded[i].CounterKind = put.CounterKind
		folded[i].CounterKindKnown = true
		folded = slices.Delete(folded, i+1, i+2)
		ctx.content.Effects = folded
		// Drop the placement clause's "them" back-reference from the sequence's
		// reference list; the exile primitive now owns the counter placement, so
		// the reference is fully consumed and must not survive as a phantom the
		// consumed-count check would report as dropped.
		ctx.content.References = referencesOutsideSpan(ctx.content.References, put.ClauseSpan)
		return ctx, true
	}
	return ctx, false
}

// exileTopNamedCounterFoldable reports whether an exile/put effect pair is the
// exact "exile ... top of your library[ face down]. Put a <name> counter on each
// of them." shape the fold can rewrite into a single counter-bearing exile.
func exileTopNamedCounterFoldable(exile, put compiler.CompiledEffect) bool {
	if exile.Kind != compiler.EffectExile ||
		exile.CardSource != parser.EffectCardSourceTopOfPlayerLibrary ||
		exile.Context != parser.EffectContextController ||
		exile.CounterKindKnown ||
		exile.Negated ||
		exile.Optional ||
		len(exile.Targets) != 0 {
		return false
	}
	if put.Kind != compiler.EffectPut ||
		put.Context != parser.EffectContextController ||
		put.Negated ||
		put.Optional ||
		!put.CounterKindKnown ||
		put.Duration != compiler.DurationNone ||
		len(put.Targets) != 0 ||
		!put.Amount.Known ||
		put.Amount.Value != 1 {
		return false
	}
	return exileTopCounterBatchReference(put.References)
}

// exileTopCounterBatchReference reports whether the placement clause's only
// reference is the plural "them" the compiler bound to the immediately preceding
// instruction's result — i.e. the just-exiled batch ("Put a <name> counter on
// each of them."). A singular, possessive, or subject pronoun, a reference bound
// to anything other than the prior instruction, or any additional reference
// denotes something other than the exiled cards and fails the fold closed. This
// prior-instruction binding, not any surface wording, is what proves the counter
// lands on the cards the exile just produced, keeping the fold text-blind.
func exileTopCounterBatchReference(references []compiler.CompiledReference) bool {
	return len(references) == 1 &&
		references[0].Kind == compiler.ReferencePronoun &&
		references[0].Pronoun == compiler.ReferencePronounThem &&
		references[0].Binding == compiler.ReferenceBindingPriorInstructionResult &&
		references[0].PriorInstruction == 0
}
