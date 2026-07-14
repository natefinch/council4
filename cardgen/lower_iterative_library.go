package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerIterativeLibraryDuplicateNameSequence lowers Tainted Pact's closed
// four-effect [Exile, may-Put, Put, Exile] shape into a single
// IterativeLibraryProcess primitive with the duplicate-name stop. The parser
// marks every effect with IterativeLibraryProcess and records the stop and
// optional-take knob on the head Exile; this text-blind lowerer reads only those
// typed fields. The body carries an intrinsic "you may" so it is reached through
// the optional-content path; any shape mismatch fails closed.
func lowerIterativeLibraryDuplicateNameSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effects := ctx.content.Effects
	if len(effects) != 4 ||
		effects[0].Kind != compiler.EffectExile ||
		effects[1].Kind != compiler.EffectPut ||
		effects[2].Kind != compiler.EffectPut ||
		effects[3].Kind != compiler.EffectExile {
		return game.AbilityContent{}, false
	}
	prim, ok := iterativeLibraryPrimitive(effects)
	if !ok || prim.Stop != game.IterativeLibraryStopDuplicateName {
		return game.AbilityContent{}, false
	}
	if !iterativeLibrarySpansCovered(ctx, effects) {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: prim}},
	}.Ability(), true
}

// lowerIterativeLibraryChosenNameSequence lowers Demonic Consultation's closed
// five-effect [Exile, Reveal, Reveal, Put, Exile] shape into a single
// IterativeLibraryProcess primitive with the chosen-name stop. The parser marks
// every effect with IterativeLibraryProcess, credits the "Choose a card name."
// prelude, and records the stop, choose-name, reveal, and pre-exile count on the
// head Exile; this text-blind lowerer reads only those typed fields. Any shape
// mismatch or optionality fails closed.
func lowerIterativeLibraryChosenNameSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effects := ctx.content.Effects
	if len(effects) != 5 ||
		effects[0].Kind != compiler.EffectExile ||
		effects[1].Kind != compiler.EffectReveal ||
		effects[2].Kind != compiler.EffectReveal ||
		effects[3].Kind != compiler.EffectPut ||
		effects[4].Kind != compiler.EffectExile {
		return game.AbilityContent{}, false
	}
	prim, ok := iterativeLibraryPrimitive(effects)
	if !ok || prim.Stop != game.IterativeLibraryStopChosenName {
		return game.AbilityContent{}, false
	}
	if !iterativeLibrarySpansCovered(ctx, effects) {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: prim}},
	}.Ability(), true
}

// iterativeLibraryPrimitive builds the IterativeLibraryProcess primitive from a
// recognized effect sequence. Every effect must carry the IterativeLibraryProcess
// marker and resolve in the controller's context, and exactly one effect (the
// head) must carry a stop predicate whose typed knobs configure the primitive.
func iterativeLibraryPrimitive(effects []compiler.CompiledEffect) (game.IterativeLibraryProcess, bool) {
	var head *compiler.CompiledEffect
	for i := range effects {
		if !effects[i].IterativeLibraryProcess ||
			effects[i].Context != parser.EffectContextController {
			return game.IterativeLibraryProcess{}, false
		}
		if effects[i].IterativeLibraryStop != parser.IterativeLibraryStopNone {
			if head != nil {
				return game.IterativeLibraryProcess{}, false
			}
			head = &effects[i]
		}
	}
	if head == nil {
		return game.IterativeLibraryProcess{}, false
	}
	stop, ok := iterativeLibraryStop(head.IterativeLibraryStop)
	if !ok {
		return game.IterativeLibraryProcess{}, false
	}
	prim := game.IterativeLibraryProcess{
		Player:       game.ControllerReference(),
		Stop:         stop,
		ChooseName:   head.IterativeLibraryChooseName,
		Reveal:       head.IterativeLibraryReveal,
		OptionalTake: head.IterativeLibraryOptionalTake,
		// The chosen-name shape (Demonic Consultation) is the only one whose
		// named card is irrelevant once matching fails, so it alone opts into the
		// absent-name sentinel that lets the player name an absent card and exile
		// the whole library. Duplicate-name effects never use a chosen name.
		AllowAbsentName: stop == game.IterativeLibraryStopChosenName,
	}
	if head.IterativeLibraryPreExile > 0 {
		prim.PreExile = game.Fixed(head.IterativeLibraryPreExile)
	}
	return prim, true
}

// iterativeLibraryStop maps the parser's stop kind to the game enum, failing
// closed for an unknown or absent predicate.
func iterativeLibraryStop(kind parser.IterativeLibraryStopKind) (game.IterativeLibraryStop, bool) {
	switch kind {
	case parser.IterativeLibraryStopChosenName:
		return game.IterativeLibraryStopChosenName, true
	case parser.IterativeLibraryStopDuplicateName:
		return game.IterativeLibraryStopDuplicateName, true
	default:
		return 0, false
	}
}

// iterativeLibrarySpansCovered reports whether every content reference and
// condition falls within one of the folded effect spans, so no reference or
// gating condition needing its own instruction is silently dropped by the
// single-primitive lowering. Tainted Pact's "unless it has the same name as
// another card exiled this way" clause compiles to a condition inside the
// may-Put span; the duplicate-name primitive subsumes that predicate, so a
// span-covered condition is expected rather than a blocker.
func iterativeLibrarySpansCovered(ctx contentCtx, effects []compiler.CompiledEffect) bool {
	spans := make([]shared.Span, len(effects))
	for i := range effects {
		spans[i] = effects[i].Span
	}
	for ri := range ctx.content.References {
		if !spanCovered(ctx.content.References[ri].Span, spans) {
			return false
		}
	}
	for ci := range ctx.content.Conditions {
		if !spanCovered(ctx.content.Conditions[ci].Span, spans) {
			return false
		}
	}
	return true
}
