package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
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
		content.ModeChoiceBonus.AdditionalMaxModes != 0 ||
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
