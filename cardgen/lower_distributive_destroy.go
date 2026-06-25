package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// destroyedForEachPlayerKey links the permanents a distributive per-player
// destroy clause destroys to the per-controller token payoff that mints one
// token for each. It is distinct from the exile-until-leaves link so the two
// distributive mechanisms never share a record set.
const destroyedForEachPlayerKey = game.LinkedKey("destroyed-for-each-player")

// lowerDestroyForEachPlayerTokenChainContent lowers the distributive Saga chapter
// "For each player, destroy up to one target creature that player controls. For
// each creature destroyed this way, its controller creates a <token>." (The Curse
// of Fenric, chapter I) into a DestroyForEachPlayer primitive paired with a
// CreateTokenForEachDestroyed payoff. The controller chooses up to one matching
// creature each player controls at resolution; the runtime links every destroyed
// permanent under destroyedForEachPlayerKey, and the payoff creates one token
// under each destroyed creature's last-known controller. The destroy clause's
// per-player target is consumed by the distributive walk rather than emitted as a
// resolving target, and the token is reconstructed from the create effect the
// parser carries the printed "<token>" wording on.
//
// It returns ok=false for any shape it does not fully consume: a non-chapter
// host, an optional, condition, mode, or keyword rider, a non-controller destroy
// context, a non-referenced-controller create context, a selector it cannot
// project, a token it cannot synthesize, or references beyond the distributive
// "that player" anchor and the per-controller "its" pronoun.
func lowerDestroyForEachPlayerTokenChainContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilityChapter ||
		ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	destroyEffect := ctx.content.Effects[0]
	if destroyEffect.Kind != compiler.EffectDestroy ||
		!destroyEffect.DestroyForEachPlayer ||
		!destroyEffect.Exact ||
		destroyEffect.Negated ||
		destroyEffect.Optional ||
		destroyEffect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	createEffect := ctx.content.Effects[1]
	if createEffect.Kind != compiler.EffectCreate ||
		!createEffect.CreateTokenForEachDestroyedThisWay ||
		!createEffect.Exact ||
		createEffect.Negated ||
		createEffect.Optional ||
		createEffect.Context != parser.EffectContextReferencedObjectController {
		return game.AbilityContent{}, false
	}
	if !referencesAreThatPlayerOrItsController(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	selection, ok := SelectionForSelector(destroyEffect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	def, ok := synthesizeCreatureTokenDef(&createEffect, nil)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: game.DestroyForEachPlayer{
					Chooser:   game.ControllerReference(),
					Selection: selection,
					LinkedKey: destroyedForEachPlayerKey,
				},
			},
			{
				Primitive: game.CreateTokenForEachDestroyed{
					Source:    game.TokenDef(def),
					LinkedKey: destroyedForEachPlayerKey,
				},
			},
		},
	}.Ability(), true
}

// referencesAreThatPlayerOrItsController reports whether every reference is the
// distributive "that player" anchor the runtime resolves per player or the "its
// controller" pronoun that names each destroyed creature's last-known controller.
// Neither names a resolving object the lowering must bind, so the distributive
// destroy consumes them in place of a target binding.
func referencesAreThatPlayerOrItsController(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind != compiler.ReferenceThatPlayer &&
			reference.Kind != compiler.ReferencePronoun {
			return false
		}
	}
	return true
}
