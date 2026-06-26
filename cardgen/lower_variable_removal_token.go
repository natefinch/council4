package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// removedTargetsForTokenKey links the permanents a variable-target removal clause
// destroys or exiles to the per-controller token payoff that mints one token for
// each. It is distinct from the per-player distributive destroy link so the two
// removal-token mechanisms never share a record set.
const removedTargetsForTokenKey = game.LinkedKey("removed-targets-for-token")

// anyNumberTargetCardinalityMax is the upper bound the parser assigns to the
// unbounded "any number of target <noun>" cardinality (effect_targets.go). It is
// the only positive signal that a target spec is the "any number" form rather
// than a small fixed plural, so the variable-removal lowering keys on it.
const anyNumberTargetCardinalityMax = 99

// maxVariableRemovalTargets bounds the "X target creatures" form's resolving
// target count at the engine's maximum legal X (maxLegalXValue), so target
// announcement enumerates a finite range while still covering every castable X.
const maxVariableRemovalTargets = 20

// lowerRemovalVariableTargetsForEachTokenContent lowers the variable-target
// removal-token family "Destroy any number of target creatures. For each creature
// destroyed this way, its controller creates a <token>." (Descent of the Dragons)
// and "Exile X target creatures. For each creature exiled this way, its
// controller creates a <token>." (Curse of the Swine) into a RemoveTargetsForToken
// primitive paired with a CreateTokenForEachDestroyed payoff. The spell announces
// one variable-count creature target spec; at resolution the runtime removes every
// chosen target as one simultaneous event, links each under removedTargetsForTokenKey,
// and the payoff mints one token under each removed creature's last-known
// controller. The "X target creatures" form binds its target count to the spell's
// chosen X via TargetSpec.CountEqualsX; the "any number" form leaves the count
// free up to the candidates on the battlefield.
//
// It returns ok=false for any shape it does not fully consume: a non-two-effect or
// targeted-elsewhere host, an optional, condition, mode, or keyword rider, a
// removal verb other than destroy or exile, a per-player distributive destroy, a
// payoff whose "<removed> this way" wording does not match the removal verb, a
// non-referenced-controller create context, a target spec it cannot project, a
// token it cannot synthesize, or a reference other than the per-controller "its"
// pronoun.
func lowerRemovalVariableTargetsForEachTokenContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	removeEffect := ctx.content.Effects[0]
	exile := removeEffect.Kind == compiler.EffectExile
	if removeEffect.Kind != compiler.EffectDestroy && !exile {
		return game.AbilityContent{}, false
	}
	if removeEffect.Negated ||
		removeEffect.Optional ||
		removeEffect.DestroyForEachPlayer ||
		removeEffect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	createEffect := ctx.content.Effects[1]
	if createEffect.Kind != compiler.EffectCreate ||
		!createEffect.Exact ||
		createEffect.Negated ||
		createEffect.Optional ||
		createEffect.Context != parser.EffectContextReferencedObjectController {
		return game.AbilityContent{}, false
	}
	// Pair the removal verb with the matching "<removed> this way" payoff so a
	// destroy clause never consumes an exiled-this-way payoff and vice versa.
	if exile {
		if !createEffect.CreateTokenForEachExiledThisWay {
			return game.AbilityContent{}, false
		}
	} else if !createEffect.CreateTokenForEachDestroyedThisWay {
		return game.AbilityContent{}, false
	}
	if !referencesAllPronoun(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	spec, ok := variableRemovalTargetSpec(ctx.content.Targets[0], removeEffect)
	if !ok {
		return game.AbilityContent{}, false
	}
	def, ok := synthesizeCreatureTokenDef(&createEffect, nil)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{spec},
		Sequence: []game.Instruction{
			{
				Primitive: game.RemoveTargetsForToken{
					Exile:     exile,
					LinkedKey: removedTargetsForTokenKey,
				},
			},
			{
				Primitive: game.CreateTokenForEachDestroyed{
					Source:    game.TokenDef(def),
					LinkedKey: removedTargetsForTokenKey,
				},
			},
		},
	}.Ability(), true
}

// variableRemovalTargetSpec projects the spell's lone creature target into a
// resolving multi-target spec. The "X target creatures" form (removeEffect carries
// the variable X on its amount) overrides the parser's single-target cardinality
// with a 0..maxVariableRemovalTargets range and sets CountEqualsX so casting is
// legal only when the chosen target count equals X. The "any number of target
// creatures" form keeps the parser's 0..anyNumberTargetCardinalityMax range. Any
// other cardinality is not a recognized variable-removal target.
func variableRemovalTargetSpec(target compiler.CompiledTarget, removeEffect compiler.CompiledEffect) (game.TargetSpec, bool) {
	countEqualsX := removeEffect.Amount.VariableX
	switch {
	case countEqualsX:
		target.Cardinality.Min = 0
		target.Cardinality.Max = maxVariableRemovalTargets
	case target.Cardinality.Min == 0 && target.Cardinality.Max == anyNumberTargetCardinalityMax:
	default:
		return game.TargetSpec{}, false
	}
	target.Exact = true
	// This lowering models the whole chosen group with one RemoveTargetsForToken
	// instruction, so it opts into the unbounded "any number of" cardinality the
	// shared builder otherwise rejects for the per-slot unroll callers.
	spec, ok := permanentTargetSpecAllowingUnbounded(target, true)
	if !ok {
		return game.TargetSpec{}, false
	}
	if countEqualsX {
		spec.CountEqualsX = true
	}
	return spec, true
}

// referencesAllPronoun reports whether every reference is the per-controller "its"
// pronoun that names each removed creature's last-known controller. That pronoun
// names no resolving object the lowering must bind, so the linked token payoff
// consumes it in place of a target binding.
func referencesAllPronoun(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind != compiler.ReferencePronoun {
			return false
		}
	}
	return true
}
