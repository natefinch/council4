package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerBecomeMonarchSpell lowers an exact monarch-designation effect (CR 720)
// to a BecomeMonarch primitive. It supports the controller form ("You become
// the monarch."), the single player-target form ("Target player becomes the
// monarch.", "Target opponent becomes the monarch."), and the triggered
// event-player form whose subject is the "that player" bound to the triggering
// event ("Whenever Starscream deals combat damage to a player, if there is no
// monarch, that player becomes the monarch."). Any other shape fails closed.
func lowerBecomeMonarchSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := contentDiagnostic(
		ctx,
		"unsupported become monarch effect",
		"the executable source backend supports only exact 'you become the monarch', single player-target become-monarch, and the triggered 'that player becomes the monarch' event-player form",
	)
	if effect.Negated || ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, unsupported
	}
	playerRef, targets, ok := becomeMonarchRecipient(ctx, effect)
	if !ok {
		return game.AbilityContent{}, unsupported
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.BecomeMonarch{Player: playerRef},
		}},
	}.Ability(), nil
}

// lowerCantBecomeMonarchSpell lowers the exact controller-scoped prohibition
// "You can't become the monarch this turn." (Jared Carthalion) to a
// CantBecomeMonarch primitive.
func lowerCantBecomeMonarchSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if effect.Negated || ctx.optional || !effect.Exact ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported can't-become-monarch effect",
			"the executable source backend supports only the exact 'You can't become the monarch this turn.' effect",
		)
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CantBecomeMonarch{Player: game.ControllerReference()},
		}},
	}.Ability(), nil
}

// becomeMonarchRecipient resolves the player who becomes the monarch from the
// effect's typed context: the resolving controller ("you"), a single named
// player target ("target player"/"target opponent"), or the "that player" bound
// to the triggering event ("that player becomes the monarch"), whose damaged
// or otherwise event-identified player the runtime resolves through
// EventPlayerReference.
func becomeMonarchRecipient(
	ctx contentCtx,
	effect compiler.CompiledEffect,
) (game.PlayerReference, []game.TargetSpec, bool) {
	switch {
	case effect.Exact && len(ctx.content.Targets) == 0 && len(ctx.content.References) == 0 &&
		effect.Context == parser.EffectContextController:
		return game.ControllerReference(), nil, true
	case effect.Exact && len(ctx.content.Targets) == 1 && effect.Context == parser.EffectContextTarget:
		spec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.PlayerReference{}, nil, false
		}
		return game.TargetPlayerReference(0), []game.TargetSpec{spec}, true
	case len(ctx.content.Targets) == 0 &&
		effect.Context == parser.EffectContextReferencedPlayer &&
		becomeMonarchEventPlayerReference(ctx.content.References):
		return game.EventPlayerReference(), nil, true
	default:
		return game.PlayerReference{}, nil, false
	}
}

// becomeMonarchEventPlayerReference reports whether the effect's references are
// exactly the lone "that player" demonstrative bound to the triggering event's
// player subject (ReferenceBindingEventPlayer), the antecedent of "that player
// becomes the monarch." That binding is established only for triggers whose
// event has an authoritative player subject (e.g. "deals combat damage to a
// player"), so the runtime EventPlayerReference resolves to the intended player.
// It fails closed for any other reference shape so no non-event-player subject
// is mistaken for the new monarch.
func becomeMonarchEventPlayerReference(references []compiler.CompiledReference) bool {
	if len(references) != 1 {
		return false
	}
	reference := references[0]
	return reference.Kind == compiler.ReferenceThatPlayer &&
		reference.Binding == compiler.ReferenceBindingEventPlayer
}
