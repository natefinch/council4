package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerBecomeColorContent lowers the one-shot continuous color-set "<subject>
// becomes <color>... until end of turn." (Cerulean Wisps, Niveous Wisps, Raging
// Spirit) into a single ApplyContinuous at LayerColor over the source or the
// single targeted permanent for the turn. The named colors SET the subject's
// color set; the colorless form clears it.
//
// The source-affecting form ("This creature becomes ...") applies to the source
// permanent and carries the inherent self reference; the targeted form ("Target
// permanent becomes ...") applies to the single target and accepts no extra
// references. Any richer shape — a negation, optional, condition, keyword, mode,
// or a target count other than the one the subject form expects — fails closed.
func lowerBecomeColorContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported color-change effect",
			"the executable source backend supports only a source or single target permanent becoming a fixed color set until end of turn",
		)
	}
	if !effect.BecomeColorUntilEndOfTurn ||
		effect.Negated ||
		effect.Optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	if effect.BecomeColorColorless == (len(effect.BecomeColorColors) != 0) {
		return unsupported()
	}
	colorEffect := game.ContinuousEffect{Layer: game.LayerColor}
	if effect.BecomeColorColorless {
		colorEffect.SetColorless = true
	} else {
		colorEffect.SetColors = slices.Clone(effect.BecomeColorColors)
	}
	continuousEffects := []game.ContinuousEffect{colorEffect}

	if effect.BecomeColorSource {
		if len(ctx.content.Targets) != 0 {
			return unsupported()
		}
		return sourceContinuousMode(continuousEffects), nil
	}
	if len(ctx.content.Targets) != 1 || len(ctx.content.References) != 0 {
		return unsupported()
	}
	return continuousTargetMode(ctx.content.Targets[0], continuousEffects, game.DurationUntilEndOfTurn, unsupported)
}
