package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// lowerOrAlternativeModal lowers a sentence-level "do A or B" body into a modal
// ability whose modes are the two alternatives ("Put a +1/+1 counter on target
// creature or that creature gains banding, first strike, or trample."). The
// controller chooses exactly one alternative at resolution, which is the modal
// "choose one" semantics realized with the existing modal machinery. Both
// alternatives act on a single shared target creature; the second alternative's
// "that creature" back-reference denotes that same target.
//
// The handled return is true once the OR shape is recognized so the caller
// commits to the modal interpretation: failing closed here keeps a body whose
// alternatives cannot both be lowered as unsupported rather than letting it fall
// through to the ordered-sequence path, which would wrongly carry out both
// alternatives instead of one.
func lowerOrAlternativeModal(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[1].Connection != parser.EffectConnectionOr {
		return game.AbilityContent{}, nil, false
	}
	unsupported := func(detail string) (game.AbilityContent, *shared.Diagnostic, bool) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported alternative effects", detail), true
	}
	if ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported("the executable source backend supports only an unconditional, non-optional two-way alternative")
	}
	if len(ctx.content.Targets) != 1 {
		return unsupported("the executable source backend supports only alternatives over a single shared target")
	}
	// Every reference must be a back-reference to the single shared target
	// (the "that creature" in the second alternative). Any other reference is a
	// shape this path does not model.
	for i := range ctx.content.References {
		if ctx.content.References[i].Binding != compiler.ReferenceBindingTarget {
			return unsupported("the executable source backend supports only back-references to the shared target across alternatives")
		}
	}
	sharedSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported("the shared alternative target is not a supported permanent target")
	}
	modesA, diagnostic := lowerOrBranchModes(cardName, ctx, 0, syntax)
	if diagnostic != nil {
		return game.AbilityContent{}, diagnostic, true
	}
	modesB, diagnostic := lowerOrBranchModes(cardName, ctx, 1, syntax)
	if diagnostic != nil {
		return game.AbilityContent{}, diagnostic, true
	}
	modes := make([]game.Mode, 0, len(modesA)+len(modesB))
	modes = append(modes, modesA...)
	modes = append(modes, modesB...)
	if len(modes) < 2 {
		return unsupported("an alternative produced no executable mode")
	}
	return game.AbilityContent{
		SharedTargets: []game.TargetSpec{sharedSpec},
		Modes:         modes,
		MinModes:      1,
		MaxModes:      1,
	}, nil, true
}

// lowerOrBranchModes lowers one alternative of an OR body in isolation and
// returns its modes with their per-mode targets stripped, since the shared
// target is declared once on the enclosing modal content. The branch effect is
// lowered as a standalone single-target effect: its connection is cleared and
// the shared target is presented directly, with the redundant back-references
// dropped. Keyword grants attach to the alternative that gains them; the other
// alternative carries no keywords.
func lowerOrBranchModes(
	cardName string,
	ctx contentCtx,
	branchIndex int,
	syntax *parser.Ability,
) ([]game.Mode, *shared.Diagnostic) {
	effect := ctx.content.Effects[branchIndex]
	effect.Connection = parser.EffectConnectionNone
	// The branch effect carried RequiresOrderedLowering only because the OR body
	// held two effects; lowered in isolation it is a genuine standalone effect,
	// so clear the flag as the other split-effect lowerers do.
	effect.RequiresOrderedLowering = false
	// A counter-removal alternative removes a counter "from it", where "it" is
	// the single shared target the put alternative places a counter on. Its
	// elided agent leaves it in a back-reference shape the standalone
	// remove-counter lowerer does not model (it expects an explicit controller
	// subject and its own target), so it is lowered directly here onto the shared
	// target. The put alternative carries an explicit subject and lowers through
	// the ordinary single-effect path.
	if effect.Kind == compiler.EffectRemoveCounter {
		mode, diagnostic := lowerOrRemoveCounterBranch(ctx, effect)
		if diagnostic != nil {
			return nil, diagnostic
		}
		return []game.Mode{mode}, nil
	}
	branchContent := ctx.content
	branchContent.Effects = []compiler.CompiledEffect{effect}
	branchContent.Targets = append([]compiler.CompiledTarget(nil), ctx.content.Targets...)
	branchContent.References = nil
	if effect.Kind == compiler.EffectGain || effect.Kind == compiler.EffectLose {
		branchContent.Keywords = append([]compiler.CompiledKeyword(nil), ctx.content.Keywords...)
	} else {
		branchContent.Keywords = nil
	}
	branchCtx := ctx
	branchCtx.content = branchContent
	content, diagnostic := lowerContent(cardName, branchCtx, syntax)
	if diagnostic != nil {
		return nil, diagnostic
	}
	if content.MinModes != 1 || content.MaxModes != 1 ||
		content.ModeChoiceBonus != (game.ModeChoiceBonus{}) ||
		content.AllowDuplicateModes ||
		len(content.Modes) == 0 {
		return nil, contentDiagnostic(ctx, "unsupported alternative effects", "an alternative lowered to an unsupported modal shape")
	}
	// The alternative must act on exactly the one shared target, declared either
	// as a per-mode target (a plain single-target effect) or as shared targets (a
	// keyword-choice grant). Reject anything that introduces additional or
	// missing targets so the flattened modal cannot silently retarget.
	if len(content.SharedTargets) > 1 {
		return nil, contentDiagnostic(ctx, "unsupported alternative effects", "an alternative declared more than one shared target")
	}
	modes := content.Modes
	for i := range modes {
		if len(modes[i].Targets) > 1 {
			return nil, contentDiagnostic(ctx, "unsupported alternative effects", "an alternative mode targeted more than one permanent")
		}
		modes[i].Targets = nil
	}
	return modes, nil
}

// lowerOrRemoveCounterBranch lowers the counter-removal alternative of a
// put-or-remove counter modal ("...or remove one from it.") into a single
// RemoveCounter instruction acting on the shared target (target permanent
// reference 0). It reads the counter kind and amount from the compiled effect:
// a named placeable kind removes that kind; the kind-elided "one" form removes
// one counter of a controller-chosen kind. The per-mode target is declared so
// the enclosing modal's shared-target stripping reattaches it, mirroring the put
// alternative. It fails closed for a negated removal, a non-positive or dynamic
// amount, a player-only or unsupported named kind, and the kind-elided plural
// that has no single-choice resolution.
func lowerOrRemoveCounterBranch(ctx contentCtx, effect compiler.CompiledEffect) (game.Mode, *shared.Diagnostic) {
	if effect.Negated ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 {
		return game.Mode{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	sharedSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.Mode{}, unsupportedCounterPlacementDiagnostic(ctx)
	}
	remove := game.RemoveCounter{
		Amount: game.Fixed(effect.Amount.Value),
		Object: game.TargetPermanentReference(0),
	}
	if effect.CounterKindKnown {
		if !compiler.CounterKindPlacementSupported(effect.CounterKind) ||
			effect.CounterKind.PlayerOnly() {
			return game.Mode{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
		remove.CounterKind = effect.CounterKind
	} else if inherited, ok := ellipticalRemovedCounterKind(ctx); ok {
		remove.CounterKind = inherited
	} else {
		if effect.Amount.Value != 1 {
			return game.Mode{}, unsupportedCounterPlacementDiagnostic(ctx)
		}
		remove.ChooseKind = true
	}
	return game.Mode{
		Targets: []game.TargetSpec{sharedSpec},
		Sequence: []game.Instruction{{
			Primitive: remove,
		}},
	}, nil
}

// ellipticalRemovedCounterKind returns the counter kind the put alternative
// places when the remove alternative elides its counter noun ("...put a lore
// counter on target Saga you control or remove one from it."). The elided "one"
// denotes a counter of the kind just placed, so the removal inherits that kind
// rather than offering the controller a choice. It returns ok=false when the put
// alternative is not a single recognized, placeable, permanent counter kind, in
// which case the caller falls back to the controller-chosen removal.
func ellipticalRemovedCounterKind(ctx contentCtx) (counter.Kind, bool) {
	if len(ctx.content.Effects) != 2 {
		return counter.Kind(0), false
	}
	put := ctx.content.Effects[0]
	if put.Kind != compiler.EffectPut ||
		!put.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(put.CounterKind) ||
		put.CounterKind.PlayerOnly() {
		return counter.Kind(0), false
	}
	return put.CounterKind, true
}
