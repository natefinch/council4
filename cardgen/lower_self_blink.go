package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// selfBlinkReturnRiders validates the self-exile-then-return two-effect body
// shared by the immediate (lowerSelfBlinkSequence) and delayed
// (lowerDelayedSelfBlinkSequence) self-blink lowerers. The exiled object is the
// source permanent itself, co-referenced by the return's "it"/"its" bound to the
// source, so this owns every check except the return's connective and delayed
// timing — which the two callers branch on to pick the immediate or delayed
// shape. It returns the validated return effect and its "with a <kind> counter on
// it" entry-counter rider, and fails closed (ok=false) for any shape it does not
// fully model.
func selfBlinkReturnRiders(ctx contentCtx) (returnEffect compiler.CompiledEffect, entryCounters []game.CounterPlacement, ok bool) {
	content := ctx.content
	if ctx.optional ||
		len(content.Effects) != 2 ||
		len(content.Targets) != 0 ||
		len(content.Conditions) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 {
		return compiler.CompiledEffect{}, nil, false
	}
	exileEffect := content.Effects[0]
	returnEffect = content.Effects[1]
	if exileEffect.Kind != compiler.EffectExile ||
		exileEffect.Negated ||
		exileEffect.Context != parser.EffectContextController ||
		exileEffect.DelayedTiming != 0 ||
		len(exileEffect.Targets) != 0 ||
		len(exileEffect.References) != 1 ||
		exileEffect.References[0].Kind != compiler.ReferenceThisObject ||
		exileEffect.References[0].Binding != compiler.ReferenceBindingSource {
		return compiler.CompiledEffect{}, nil, false
	}
	if returnEffect.Kind != compiler.EffectReturn ||
		returnEffect.Negated ||
		returnEffect.ToZone != zone.Battlefield ||
		returnEffect.EntersColorChoice ||
		returnEffect.EntersTypeChoice ||
		returnEffect.EntersWithCounters ||
		len(returnEffect.Targets) != 0 ||
		len(returnEffect.References) == 0 {
		return compiler.CompiledEffect{}, nil, false
	}
	// Every effect reference is consumed below; the content-level reference list
	// must hold exactly the exile's "this creature" plus the return's "it"/"its"
	// so nothing is silently dropped.
	if len(content.References) != len(exileEffect.References)+len(returnEffect.References) {
		return compiler.CompiledEffect{}, nil, false
	}
	// The return's "it"/"its"/"this creature" co-reference the just-exiled source
	// permanent; one of them must name it directly ("it" or "this creature") so
	// the clause carries a return object.
	hasDirectObject := false
	for _, ref := range returnEffect.References {
		if ref.Binding != compiler.ReferenceBindingSource {
			return compiler.CompiledEffect{}, nil, false
		}
		switch {
		case ref.Kind == compiler.ReferenceThisObject:
			hasDirectObject = true
		case ref.Kind == compiler.ReferencePronoun && ref.Pronoun == compiler.ReferencePronounIt:
			hasDirectObject = true
		case ref.Kind == compiler.ReferencePronoun && ref.Pronoun == compiler.ReferencePronounIts:
		default:
			return compiler.CompiledEffect{}, nil, false
		}
	}
	if !hasDirectObject {
		return compiler.CompiledEffect{}, nil, false
	}
	entryCounters, ok = blinkEntryCounters(returnEffect)
	if !ok {
		return compiler.CompiledEffect{}, nil, false
	}
	return returnEffect, entryCounters, true
}

// selfBlinkPutOnBattlefield builds the put-onto-battlefield instruction that
// returns the linked self-exiled permanent, carrying the validated "tapped" entry
// rider, the fixed entry counters, and the "under your control" controller rider.
func selfBlinkPutOnBattlefield(
	key game.LinkedKey,
	returnEffect compiler.CompiledEffect,
	entryCounters []game.CounterPlacement,
) game.PutOnBattlefield {
	put := game.PutOnBattlefield{
		Source:           game.LinkedBattlefieldSource(key),
		EntryTapped:      returnEffect.EntersTapped,
		EntryTransformed: returnEffect.EntersTransformed,
		EntryCounters:    entryCounters,
	}
	if returnEffect.UnderYourControl {
		put.Recipient = opt.Val(game.ControllerReference())
	}
	return put
}

// lowerDelayedSelfBlinkSequence lowers the delayed self-blink "Exile this
// creature. Return it to the battlefield [tapped] under [its owner's|your]
// control at the beginning of the next end step." (Argent Sphinx, Saltskitter,
// Anurid Brushhopper, Ghost Council of Orzhova). It shares the self-exile-then-
// return contract with the immediate lowerSelfBlinkSequence but the return is a
// separate sentence whose delayed timing wraps the put-onto-battlefield in a
// delayed trigger, mirroring lowerDelayedBlinkReturn for the single-target form.
// It fails closed for any shape it does not fully model.
func lowerDelayedSelfBlinkSequence(ctx contentCtx) (game.AbilityContent, bool) {
	returnEffect, entryCounters, ok := selfBlinkReturnRiders(ctx)
	if !ok || returnEffect.DelayedTiming != game.DelayedAtBeginningOfNextEndStep {
		return game.AbilityContent{}, false
	}
	key := game.LinkedKey("delayed-self-blink")
	exile := game.Exile{Object: game.SourcePermanentReference(), ExileLinkedKey: key}
	put := selfBlinkPutOnBattlefield(key, returnEffect, entryCounters)
	delayed := game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
		Timing:  game.DelayedAtBeginningOfNextEndStep,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: put}}}.Ability(),
	}}
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: exile},
		{Primitive: delayed},
	}}.Ability(), true
}

// lowerImmediateSelfBlinkReturn lowers the immediate self-blink return clause of
// an optional "you may exile this <permanent>. If you do, return it to the
// battlefield [tapped] under [its owner's|your] control." trigger (Estrid's
// Invocation's granted upkeep ability). Unlike lowerSelfBlinkSequence, whose
// ", then return …" connective folds the whole two-effect body at once, the "If
// you do, …" wording makes the return a separate sentence gated on the optional
// self-exile, so it arrives as a standalone clause after the exile has already
// been lowered into sequence. This mirrors lowerImmediateBlinkReturn for the
// single-target flicker, but the exiled object is the source permanent itself and
// the return's "it"/"its" co-references the source rather than a prior
// instruction's target result. It rewrites the preceding self-exile instruction
// to remember the exiled card under a linked key and returns the
// put-onto-battlefield content; the optional-flow envelope then gates that put on
// the exile succeeding. It fails closed for any shape it does not fully model.
func lowerImmediateSelfBlinkReturn(
	effects []compiler.CompiledEffect,
	effectIndex int,
	ctx contentCtx,
	sequence []game.Instruction,
) (game.Exile, game.AbilityContent, bool) {
	// Invariant: ctx is the contextForEffect-narrowed per-clause effectAbility
	// built at lowerOrderedEffectSequence and threaded through
	// lowerDelayedSequenceClause, so content.Effects always holds exactly one
	// clause effect; any other length is an upstream bug.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerImmediateSelfBlinkReturn: expected a single effect, got %d", len(ctx.content.Effects)))
	}
	returnEffect := ctx.content.Effects[0]
	if effectIndex == 0 ||
		len(sequence) != effectIndex ||
		effects[effectIndex-1].Kind != compiler.EffectExile ||
		effects[effectIndex-1].DelayedTiming != 0 ||
		returnEffect.Kind != compiler.EffectReturn ||
		// Only an immediate return lowers here; a delayed return keeps its
		// next-end-step timing and is wrapped in a delayed trigger elsewhere.
		returnEffect.DelayedTiming != 0 ||
		returnEffect.Negated ||
		returnEffect.ToZone != zone.Battlefield ||
		returnEffect.EntersColorChoice ||
		returnEffect.EntersTypeChoice ||
		returnEffect.EntersWithCounters ||
		len(returnEffect.References) == 0 {
		return game.Exile{}, game.AbilityContent{}, false
	}
	// The return's "it"/"its"/"this <permanent>" co-reference the just-exiled
	// source permanent; one of them must name it directly ("it" or "this
	// <permanent>") so the clause carries a return object.
	hasDirectObject := false
	for _, ref := range returnEffect.References {
		if ref.Binding != compiler.ReferenceBindingSource {
			return game.Exile{}, game.AbilityContent{}, false
		}
		switch {
		case ref.Kind == compiler.ReferenceThisObject:
			hasDirectObject = true
		case ref.Kind == compiler.ReferencePronoun && ref.Pronoun == compiler.ReferencePronounIt:
			hasDirectObject = true
		case ref.Kind == compiler.ReferencePronoun && ref.Pronoun == compiler.ReferencePronounIts:
		default:
			return game.Exile{}, game.AbilityContent{}, false
		}
	}
	if !hasDirectObject {
		return game.Exile{}, game.AbilityContent{}, false
	}
	// "with a <kind> counter on it" rider: only fixed, known, positive counts of a
	// known kind are modeled; every other counter form fails closed.
	var entryCounters []game.CounterPlacement
	if returnEffect.CounterKindKnown {
		if !returnEffect.Amount.Known || returnEffect.Amount.Value < 1 {
			return game.Exile{}, game.AbilityContent{}, false
		}
		entryCounters = []game.CounterPlacement{{
			Kind:   returnEffect.CounterKind,
			Amount: returnEffect.Amount.Value,
		}}
	}
	// References validated — clear before fail-closed check.
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.Exile{}, game.AbilityContent{}, false
	}
	exile, ok := sequence[effectIndex-1].Primitive.(game.Exile)
	if !ok ||
		exile.Group.Valid() ||
		(exile.Object != game.SourcePermanentReference() && exile.Object != game.SourceCardPermanentReference()) ||
		exile.ExileLinkedKey != "" {
		return game.Exile{}, game.AbilityContent{}, false
	}
	key := game.LinkedKey(fmt.Sprintf("self-blink-%d", effectIndex))
	exile.ExileLinkedKey = key
	put := selfBlinkPutOnBattlefield(key, returnEffect, entryCounters)
	return exile, game.Mode{Sequence: []game.Instruction{{Primitive: put}}}.Ability(), true
}
