package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerExileForEachPlayerUntilLeavesContent lowers the distributive Saga chapter
// "For each player, exile up to one [other] target <permanent> that player
// controls until this Saga leaves the battlefield." (Vault 13: Dweller's
// Journey) into an ExileForEachPlayer primitive. The controller chooses up to
// one matching permanent each player controls at resolution and the runtime
// links every exiled permanent under exileUntilLeavesKey, so the paired chapter
// return — and the face-level synthesized leaves-the-battlefield safety net —
// release exactly this set. The candidate filter is reconstructed from the
// effect selector the parser carries the printed "[other] <permanent>" wording
// on; the per-player "that player controls" scope is applied by the runtime
// rather than a selector controller predicate.
//
// It returns ok=false for any shape it does not fully consume: a non-chapter
// host, an optional, condition, mode, or keyword rider, a non-controller
// context, a target it cannot drop, a selector it cannot project, or references
// beyond the distributive "that player" anchor and the source duration anchor.
func lowerExileForEachPlayerUntilLeavesContent(ctx contentCtx) (game.AbilityContent, bool) {
	// lowerContent calls this only from its len(Effects)==1 block, so a different
	// effect count is a dispatch bug rather than an unsupported card.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf("lowerExileForEachPlayerUntilLeavesContent: reached with %d effects; lowerContent dispatches here only for single-effect content", len(ctx.content.Effects)))
	}
	if ctx.enclosingKind != compiler.AbilityChapter ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectExile ||
		!effect.ExileForEachPlayerUntilSourceLeaves ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	if !referencesAreSourceAnchorsOrThatPlayer(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	selection, ok := SelectionForSelector(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ExileForEachPlayer{
				Chooser:   game.ControllerReference(),
				Selection: selection,
				LinkedKey: exileUntilLeavesKey,
			},
		}},
	}.Ability(), true
}

// referencesAreSourceAnchorsOrThatPlayer reports whether every reference is
// either the source duration anchor ("until this Saga leaves the battlefield")
// or the distributive "that player" anchor that the runtime resolves per player.
// Neither names a resolving object the distributive exile must bind, so the
// lowering consumes them in place of a target binding.
func referencesAreSourceAnchorsOrThatPlayer(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind == compiler.ReferenceThatPlayer {
			continue
		}
		if reference.Binding != compiler.ReferenceBindingSource {
			return false
		}
		if reference.Kind != compiler.ReferenceThisObject &&
			reference.Kind != compiler.ReferenceSelfName {
			return false
		}
	}
	return true
}

// lowerReturnLinkedExiledPartialContent lowers the Saga chapter payoff "Return
// <count> cards exiled with this Saga to the battlefield under their owners'
// control and put the rest on the bottom of their owners' libraries." (Vault 13:
// Dweller's Journey) into a ReturnLinkedExiledCardsToBattlefield primitive. The
// controller chooses <count> of the cards a sibling distributive exile recorded
// under exileUntilLeavesKey to return under their owners' control; the unreturned
// remainder goes to the bottom of their owners' libraries. The cards are
// identified through the source link rather than targets, so the clause carries
// no target.
//
// It returns ok=false for any shape it does not fully consume: a non-chapter
// host, an optional, condition, mode, keyword, or target rider, a non-controller
// context, a count that is not a fixed positive number, or a missing remainder
// disposal clause.
func lowerReturnLinkedExiledPartialContent(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilityChapter ||
		ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	returnEffect := ctx.content.Effects[0]
	if returnEffect.Kind != compiler.EffectReturn ||
		!returnEffect.ReturnLinkedExiledToBattlefieldPartial ||
		!returnEffect.Exact ||
		returnEffect.Negated ||
		returnEffect.Optional ||
		returnEffect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	if !returnEffect.Amount.Known ||
		returnEffect.Amount.RangeKnown ||
		returnEffect.Amount.VariableX ||
		returnEffect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		returnEffect.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	putEffect := ctx.content.Effects[1]
	if putEffect.Kind != compiler.EffectPut ||
		!putEffect.PutLinkedExiledRestOnLibraryBottom ||
		!putEffect.Exact ||
		putEffect.Negated ||
		putEffect.Optional ||
		putEffect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ReturnLinkedExiledCardsToBattlefield{
				Chooser:             game.ControllerReference(),
				LinkedKey:           exileUntilLeavesKey,
				Amount:              game.Fixed(returnEffect.Amount.Value),
				RestToLibraryBottom: true,
			},
		}},
	}.Ability(), true
}
