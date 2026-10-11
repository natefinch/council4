package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// playerRecipient owns only the actor and any newly announced targets. Amount
// operands and verb-specific payloads must be validated separately.
type playerRecipient struct {
	player  game.PlayerReference
	group   game.PlayerGroupReference
	targets []game.TargetSpec
}

func (r playerRecipient) mode(
	amount game.Quantity,
	single func(game.Quantity, game.PlayerReference) game.Primitive,
	group func(game.Quantity, game.PlayerGroupReference) game.Primitive,
) game.Mode {
	var primitive game.Primitive
	if r.group != (game.PlayerGroupReference{}) {
		primitive = group(amount, r.group)
	} else {
		primitive = single(amount, r.player)
	}
	return game.Mode{
		Targets: r.targets,
		Sequence: []game.Instruction{{
			Primitive: primitive,
		}},
	}
}

// resolvePlayerRecipient consumes the recipient references remaining after the
// caller has validated its amount operands. It never infers a missing actor
// from an inherited object or treats an owner as that object's controller.
func resolvePlayerRecipient(ctx contentCtx, references []compiler.CompiledReference) (playerRecipient, bool) {
	if len(ctx.content.Effects) != 1 {
		return playerRecipient{}, false
	}
	effect := ctx.content.Effects[0]
	targets := ctx.content.Targets
	switch effect.Context {
	case parser.EffectContextController:
		if len(targets) == 0 && len(references) == 0 {
			return playerRecipient{player: game.ControllerReference()}, true
		}
	case parser.EffectContextDefendingPlayer:
		if len(targets) == 0 && len(references) == 0 {
			return playerRecipient{player: game.DefendingPlayerReference()}, true
		}
	case parser.EffectContextEachOpponent, parser.EffectContextEachOtherPlayer:
		if len(targets) == 0 && len(references) == 0 {
			return playerRecipient{group: game.OpponentsReference()}, true
		}
	case parser.EffectContextEachPlayer:
		if len(targets) == 0 && len(references) == 0 {
			return playerRecipient{group: game.AllPlayersReference()}, true
		}
	case parser.EffectContextEachOpponentDealtCombatDamageByNamed:
		if len(targets) == 0 && len(references) == 0 && effect.CombatDamageSourceName != "" {
			return playerRecipient{
				group: game.OpponentsDealtCombatDamageThisGameByNamedReference(effect.CombatDamageSourceName),
			}, true
		}
	case parser.EffectContextTarget, parser.EffectContextPriorSubject:
		if len(targets) == 1 && len(references) == 0 {
			spec, ok := playerTargetSpec(targets[0])
			if ok {
				return playerRecipient{
					player:  game.TargetPlayerReference(0),
					targets: []game.TargetSpec{spec},
				}, true
			}
		}
	case parser.EffectContextReferencedObjectController, parser.EffectContextReferencedObjectOwner:
		recipientCtx := ctx
		recipientCtx.content.References = references
		if len(targets) == 1 {
			player, ok := referencedObjectPlayerRef(recipientCtx)
			return playerRecipient{player: player}, ok
		}
		if len(targets) == 0 && effect.Context == parser.EffectContextReferencedObjectController {
			player, ok := eventReferencedObjectControllerPlayer(references)
			return playerRecipient{player: player}, ok
		}
	case parser.EffectContextEventPlayer, parser.EffectContextReferencedPlayer:
		if len(references) != 1 {
			break
		}
		ref := references[0]
		if len(targets) == 0 {
			if projectedEventControllerPlayer(ctx) && ctx.content.OwnsSubject(ref) {
				return playerRecipient{player: game.ObjectControllerReference(game.EventPermanentReference())}, true
			}
			player, ok := lowerPlayerReference(ref, referenceLoweringContext{AllowEvent: true})
			return playerRecipient{player: player}, ok
		}
		if len(targets) == 1 && effect.Context == parser.EffectContextEventPlayer &&
			ref.Kind == compiler.ReferencePronoun && ref.Pronoun == compiler.ReferencePronounThey &&
			ref.Binding == compiler.ReferenceBindingTarget && ref.Occurrence >= 0 {
			if _, ok := playerTargetSpec(targets[0]); ok {
				return playerRecipient{player: game.TargetPlayerReference(ref.Occurrence)}, true
			}
		}
		if len(targets) == 1 && effect.Context == parser.EffectContextReferencedPlayer &&
			ref.Kind == compiler.ReferenceThatPlayer &&
			ref.Binding == compiler.ReferenceBindingTarget && ref.Occurrence >= 0 {
			player, ok := referencedThatPlayerRef(targets[0], ref.Occurrence)
			return playerRecipient{player: player}, ok
		}
	default:
	}
	return playerRecipient{}, false
}

// referencedObjectPlayerRef uses the antecedent selector to retain permanent
// versus stack identity and the reference's clause-local target occurrence.
func referencedObjectPlayerRef(ctx contentCtx) (game.PlayerReference, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.References) != 1 ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingTarget ||
		ctx.content.References[0].Occurrence < 0 ||
		len(ctx.content.Targets) != 1 {
		return game.PlayerReference{}, false
	}
	object, ok := inheritedRemovalTargetObjectRef(ctx.content.Targets[0], ctx.content.References[0].Occurrence)
	if !ok {
		return game.PlayerReference{}, false
	}
	switch ctx.content.Effects[0].Context {
	case parser.EffectContextReferencedObjectController:
		return game.ObjectControllerReference(object), true
	case parser.EffectContextReferencedObjectOwner:
		return game.ObjectOwnerReference(object), true
	default:
		return game.PlayerReference{}, false
	}
}

func referencedThatPlayerRef(target compiler.CompiledTarget, occurrence int) (game.PlayerReference, bool) {
	if target.Selector.Kind == compiler.SelectorCard &&
		target.Selector.Zone == zone.Graveyard &&
		target.Selector.Controller == compiler.ControllerOpponent {
		return game.ObjectOwnerReference(game.TargetCardReference(occurrence)), true
	}
	switch target.Selector.Kind {
	case compiler.SelectorPlayer, compiler.SelectorOpponent:
		return game.TargetPlayerReference(occurrence), true
	default:
	}
	object, ok := inheritedRemovalTargetObjectRef(target, occurrence)
	if !ok {
		return game.PlayerReference{}, false
	}
	return game.ObjectControllerReference(object), true
}
