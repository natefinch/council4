package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// applyAnotherTargetDistinctness marks a later "another target ..." spec distinct
// from the spell's earlier targets (CR 601.2c, CR 115.3). When a spell announces
// more than one target and a later one reads "another"/"other" target, that word
// means the target must differ from the previously announced target, so the spec
// carries DistinctFromPriorTargets. This is the cross-target distinctness the
// runtime enforces at announcement across both the object (permanent) and card
// numbering domains.
//
// The self-exclusion a bare "another"/"other" selector otherwise lowers to
// (Selection.ExcludeSource, "another creature you control" excludes the source
// permanent) is not cross-target distinctness, so it is cleared for such a spec:
// a spell is not a permanent, and the word refers to the earlier target, not the
// source. Only a target that follows an earlier target (game slot > 0) is
// affected; the first target and any spell with no "another" target are returned
// unchanged, so every card without a genuine later-clause "another target" stays
// byte-identical.
func applyAnotherTargetDistinctness(
	targets []game.TargetSpec,
	compiled []compiler.CompiledTarget,
	spanToIdx map[shared.Span]int,
) []game.TargetSpec {
	for i := range compiled {
		if !compiled[i].Selector.Another && !compiled[i].Selector.Other {
			continue
		}
		slot, ok := spanToIdx[compiled[i].Span]
		if !ok || slot <= 0 || slot >= len(targets) {
			continue
		}
		targets[slot].DistinctFromPriorTargets = true
		clearTargetExcludeSource(&targets[slot])
	}
	return targets
}

// clearTargetExcludeSource drops a spec's Selection.ExcludeSource, collapsing the
// Selection back to absent when nothing else remains. It is used when a later
// "another target" spec's self-exclusion is superseded by cross-target
// distinctness.
func clearTargetExcludeSource(spec *game.TargetSpec) {
	if !spec.Selection.Exists {
		return
	}
	selection := spec.Selection.Val
	if !selection.ExcludeSource {
		return
	}
	selection.ExcludeSource = false
	if selection.Empty() {
		spec.Selection = opt.V[game.Selection]{}
		return
	}
	spec.Selection = opt.Val(selection)
}
