package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// resolveReanimationAura resolves a graveyard-reanimation Aura spell (Animate
// Dead, Dance of the Dead) after its Aura permanent has entered the
// battlefield. The Aura's single target is a creature card in a graveyard; on
// resolution the Aura returns that card to the battlefield under the Aura's
// controller and attaches itself to the returned permanent in one step.
//
// Doing the return-and-attach inline during resolution — rather than through a
// separate enters-the-battlefield triggered ability that uses the stack — is
// what keeps the Aura from being sent to the graveyard by the unattached-Aura
// state-based action (CR 704.5m): no player receives priority, and therefore no
// state-based action is checked, between the Aura entering and it becoming
// attached to the creature it returns.
//
// After the return, the Aura's attachment legality transitions from "creature
// card in a graveyard" to "the specific permanent it put onto the battlefield"
// (Permanent.ReanimationLinkedObject); see auraCanAttachToPermanent. The
// returned creature is also remembered under game.ReanimationLinkID so the
// Aura's leaves-the-battlefield triggered ability can sacrifice that same
// permanent.
func (e *Engine) resolveReanimationAura(g *game.Game, obj *game.StackObject, aura *game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	r := newEffectResolver(e, g, obj, agents, log)
	key := linkedObjectSourceKey(g, obj, game.ReanimationLinkID)
	clearLinkedObjects(g, key)

	creature, ok := r.putReferencedCardOnBattlefieldValue(
		game.CardReference{Kind: game.CardReferenceTarget},
		game.ControllerReference(),
		nil,
		permanentCreationOptions{},
		false,
	)
	if !ok || creature == nil {
		// The enchanted card could not be returned to the battlefield (for
		// example a replacement effect redirected it, or a commander's owner put
		// it into the command zone instead). The Aura has nothing to enchant, so
		// it is put into its owner's graveyard exactly as the unattached-Aura
		// state-based action would have done.
		movePermanentToZone(g, aura, zone.Graveyard)
		return "graveyard"
	}

	aura.ReanimationLinkedObject = creature.ObjectID
	rememberLinkedObject(g, key, game.LinkedObjectRef{ObjectID: creature.ObjectID, CardID: creature.CardInstanceID})

	if !attachPermanent(g, aura, creature) {
		// Attachment legality is granted by ReanimationLinkedObject, so this is
		// unreachable in practice; fail closed by clearing the link and sending
		// the now-unattachable Aura to the graveyard rather than leaving it in an
		// inconsistent state.
		aura.ReanimationLinkedObject = 0
		clearLinkedObjects(g, key)
		movePermanentToZone(g, aura, zone.Graveyard)
		return "graveyard"
	}
	return "battlefield"
}
