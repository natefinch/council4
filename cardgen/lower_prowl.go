package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerExilePermanentForPlay lowers Prowl, Stoic Strategist's attack trigger
// "exile up to one other target tapped creature or Vehicle. For as long as that
// card remains exiled, its owner may play it." into a single
// ExilePermanentForPlay primitive.
//
// The compiler models the two sentences as two effects sharing the ability's
// single up-to-one target: effect[0] exiles the target permanent from the
// battlefield, and effect[1] grants the exiled card's OWNER permission to play
// it for as long as it remains exiled (EffectPlay with the referenced-object-
// owner recipient and the for-as-long-as-exiled duration). Because exiling the
// permanent moves the card to its owner's exile and grants a duration-scoped
// play permission bound to that same card, the move and the owner-scoped grant
// cannot be two independent instructions; the combined primitive performs both
// atomically and remembers the exiled card under the shared self-exile link so
// the paired "whenever a player plays a card exiled with Prowl" trigger
// recognizes its provenance.
//
// Any other shape (a different trigger event, a non-owner or non-while-exiled
// grant, extra effects, modes, or conditions) fails closed and flows through
// generic lowering.
func lowerExilePermanentForPlay(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.enclosingKind != compiler.AbilityTriggered ||
		ctx.triggerEvent != game.EventAttackerDeclared ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	exile := ctx.content.Effects[0]
	if exile.Kind != compiler.EffectExile ||
		exile.Negated ||
		exile.Optional ||
		exile.Context != parser.EffectContextController ||
		exile.Duration != compiler.DurationNone ||
		exile.FromZone != zone.None {
		return game.AbilityContent{}, false
	}
	grant := ctx.content.Effects[1]
	if grant.Kind != compiler.EffectPlay ||
		grant.Negated ||
		!grant.Optional ||
		grant.Context != parser.EffectContextReferencedObjectOwner ||
		grant.Duration != compiler.DurationForAsLongAsExiled {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := exilePermanentForPlayTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.ExilePermanentForPlay{
				Object:    game.TargetPermanentReference(0),
				LinkedKey: selfExileLinkKey,
			},
		}},
	}.Ability(), true
}

// exilePermanentForPlayTargetSpec builds the TargetSpec for Prowl's exile target
// "up to one other target tapped creature or Vehicle": an up-to-one, source-
// excluding, tapped, cross-dimension (card type or artifact subtype) union.
//
// The bare cross-dimension union lowers through alternativePermanentTargetSpec
// (a Selection.AnyOf over the two members), but that shared path does not itself
// apply the union's outer tapped-state or "other" exclusion. Strip those outer
// qualifiers from a copy of the target, build the union selection, then re-apply
// them to the combined selection: the runtime evaluates a top-level Tapped and
// ExcludeSource together with AnyOf (an intersection), so the reassembled
// selection means "a tapped permanent that is not the source and is a creature
// or a Vehicle". Any other outer combat state stays rejected by
// alternativePermanentTargetSpec, so an unsupported wording fails closed. Only
// this Prowl lowering calls it, so the shared target exactness for the other
// single-object union verbs is unaffected.
func exilePermanentForPlayTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	selector := target.Selector
	if selector.Kind != compiler.SelectorUnknown || len(selector.Alternatives) == 0 {
		return game.TargetSpec{}, false
	}
	if target.Cardinality.Max < 1 || target.Cardinality.Min < 0 ||
		target.Cardinality.Min > target.Cardinality.Max {
		return game.TargetSpec{}, false
	}
	tapped := selector.Tapped
	untapped := selector.Untapped
	excludeSource := selector.Another || selector.Other
	stripped := target
	stripped.Selector.Tapped = false
	stripped.Selector.Untapped = false
	stripped.Selector.Another = false
	stripped.Selector.Other = false
	stripped.Exact = true
	spec := game.TargetSpec{
		MinTargets: target.Cardinality.Min,
		MaxTargets: target.Cardinality.Max,
		Allow:      game.TargetAllowPermanent,
	}
	built, ok := alternativePermanentTargetSpec(&stripped, &spec, true)
	if !ok {
		return game.TargetSpec{}, false
	}
	selection := built.Selection.Val
	if tapped {
		selection.Tapped = game.TriTrue
	} else if untapped {
		selection.Tapped = game.TriFalse
	}
	if excludeSource {
		selection.ExcludeSource = true
	}
	built.Selection = opt.Val(selection)
	return built, true
}
