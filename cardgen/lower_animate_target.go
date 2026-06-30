package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerAnimateTargetContent lowers the one-shot continuous target-animation
// "[Until end of turn,] target land becomes a N/N [<color>...] <subtype>...
// creature [with <keyword>...] until end of turn [that's still a land]." (Animate
// Land, Vivify, Hydroform, Kamahl, Soilshaper, Lifespark Spellbomb) into an
// ApplyContinuous over the single targeted land for the turn.
//
// The continuous effects span the layers the animation touches; they are built by
// the shared animationContinuousEffects helper, which the self-animation lowerer
// also uses. It mirrors lowerAnimateSelfContent but binds the effects to a target
// slot rather than the source. Any richer shape — no target, an extra condition,
// keyword, mode, or reference, or an unsupported animated color or keyword — fails
// closed.
func lowerAnimateTargetContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	payload := effect.AnimateTarget
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported target-animation effect",
			"the executable source backend supports only a single target land becoming a fixed N/N creature until end of turn",
		)
	}
	if payload == nil {
		return unsupported()
	}
	if effect.Negated ||
		effect.Optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Targets) != 1 {
		return unsupported()
	}

	continuousEffects, ok := animationContinuousEffects(payload)
	if !ok {
		return unsupported()
	}
	return continuousTargetMode(ctx.content.Targets[0], continuousEffects, game.DurationUntilEndOfTurn, unsupported)
}
