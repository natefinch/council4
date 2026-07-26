package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleMovePermanent resolves the single "move a permanent to a zone"
// primitive. It replaces handleBounce, handleExile, and
// handlePutPermanentOnLibrary, which were the same three code paths (controlled
// choice, whole group, single object) written out once per destination zone.
//
// The destination now flows through as data, so a wording that combines an
// existing selection with a different zone resolves with no new runtime code.
func handleMovePermanent(r *effectResolver, prim game.MovePermanent) effectResolved {
	res := effectResolved{accepted: true}

	var linkedKey game.LinkedObjectKey
	if prim.PublishLinked != "" {
		linkedKey = linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		clearLinkedObjects(r.game, linkedKey)
	}

	if prim.ControlledChoice {
		chosen := r.chooseControlledMovePermanents(prim)
		res.succeeded = movePermanentsToZoneSimultaneously(r.game, chosen, prim.Destination)
		if res.succeeded {
			r.placeMovedOnLibraryBottom(prim, chosen)
		}
		return res
	}

	targets := r.resolveObjectGroup(prim.Object, prim.Group)
	if !targets.single {
		// A group move that carries a linked key (group blink) must remember
		// every permanent it moved under that key, capturing each link before
		// the move so a later linked return brings the whole group back
		// together.
		refs := make([]game.LinkedObjectRef, len(targets.permanents))
		for i, permanent := range targets.permanents {
			refs[i] = permanentLinkedObjectRef(permanent)
		}
		moved := make([]*game.Permanent, 0, len(targets.permanents))
		for i, result := range movePermanentsToZoneSimultaneouslyWithResults(r.game, targets.permanents, prim.Destination) {
			if !result.moved {
				continue
			}
			res.succeeded = true
			moved = append(moved, result.permanent)
			if prim.PublishLinked != "" && i < len(refs) {
				rememberLinkedObject(r.game, linkedKey, refs[i])
			}
		}
		r.placeMovedOnLibraryBottom(prim, moved)
		return res
	}

	if targets.resolved {
		permanent := targets.permanents[0]
		linkedObjectRef := permanentLinkedObjectRef(permanent)
		res.succeeded = movePermanentToZone(r.game, permanent, prim.Destination)
		if res.succeeded {
			r.placeMovedOnLibraryBottom(prim, targets.permanents)
		}
		if prim.PublishLinked != "" {
			// The source may itself have changed object identity during the
			// move (a permanent exiling itself), so re-derive the key and clear
			// the new one before publishing under it.
			postMoveKey := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
			if postMoveKey != linkedKey {
				clearLinkedObjects(r.game, postMoveKey)
			}
			rememberLinkedObject(r.game, postMoveKey, linkedObjectRef)
		}
		return res
	}

	// An unresolved object reference may still name a spell on the stack, which
	// backs "return target spell to its owner's hand".
	if prim.Destination == zone.Hand {
		if resolved, ok := resolveObjectReference(r.game, r.obj, prim.Object); ok && resolved.stack != nil {
			res.succeeded = bounceStackSpellToHand(r.game, resolved.stack)
		}
	}
	return res
}

// placeMovedOnLibraryBottom moves cards that just reached their owners'
// libraries from the top to the bottom. movePermanentToZone always places a card
// on top, so the bottom wording is a follow-up reorder. Tokens are skipped: they
// cease to exist rather than becoming library cards.
func (r *effectResolver) placeMovedOnLibraryBottom(prim game.MovePermanent, moved []*game.Permanent) {
	if !prim.LibraryBottom {
		return
	}
	for _, permanent := range moved {
		if permanent == nil || permanent.Token {
			continue
		}
		player, ok := playerByID(r.game, permanent.Owner)
		if !ok {
			continue
		}
		if player.Library.Remove(permanent.CardInstanceID) {
			player.Library.AddToBottom(permanent.CardInstanceID)
		}
	}
}

// chooseControlledMovePermanents has the resolving controller choose prim.Amount
// permanents from prim.Group's candidate pool, for "Return a creature you
// control to its owner's hand." style effects. When the candidate pool holds no
// more permanents than the requested amount, every candidate is chosen without a
// prompt.
func (r *effectResolver) chooseControlledMovePermanents(prim game.MovePermanent) []*game.Permanent {
	amount := r.quantity(prim.Amount)
	if amount <= 0 {
		return nil
	}
	candidates := r.groupPermanents(prim.Group)
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) <= amount {
		return candidates
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(r.game, permanent), Card: permanentChoiceInfo(r.game, permanent)}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           r.obj.Controller,
		Prompt:           movePermanentChoicePrompt(prim.Destination),
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}
	selected := r.engine.chooseChoice(r.game, r.agents, request, r.log)
	chosen := make([]*game.Permanent, 0, len(selected))
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			chosen = append(chosen, candidates[idx])
		}
	}
	return chosen
}

// movePermanentChoicePrompt names the destination in the chooser's prompt so a
// generalized move reads like the card that produced it.
func movePermanentChoicePrompt(destination zone.Type) string {
	switch destination {
	case zone.Hand:
		return "Choose a permanent to return to its owner's hand"
	case zone.Exile:
		return "Choose a permanent to exile"
	case zone.Library:
		return "Choose a permanent to put into its owner's library"
	case zone.Graveyard:
		return "Choose a permanent to put into its owner's graveyard"
	default:
		return "Choose a permanent to move"
	}
}

// handleMoveResolvingSpell redirects the resolving spell away from its owner's
// graveyard. It replaces handleExile's SourceSpell branch and
// handleShuffleSpellIntoLibrary, which set two different StackObject flags from
// two different primitives.
func handleMoveResolvingSpell(r *effectResolver, prim game.MoveResolvingSpell) effectResolved {
	res := effectResolved{accepted: true}
	if r.obj == nil {
		return res
	}
	switch prim.Destination {
	case zone.Exile:
		r.obj.ExileOnResolution = true
	case zone.Library:
		r.obj.ShuffleIntoLibraryOnResolution = true
	default:
		// Unreachable: validatePrimitive rejects every other destination.
		return res
	}
	res.succeeded = true
	return res
}
