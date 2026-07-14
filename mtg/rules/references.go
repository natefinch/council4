package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// referenceResolver is the internal module that binds a game and a resolving
// stack object and owns every runtime reference lookup: target-slot objects and
// players, source and attached permanents, linked objects, event/last-known
// information, alive-player checks, and group enumeration with Selection
// matching and exclusions. It is created per resolution; the resolveObjectReference,
// resolvePlayerReference, permanentAt, and playerAt entry points are thin
// adapters over it.
//
// By default the resolver derives the source permanent from the stack object
// (obj.SourceID). A caller whose source basis differs from the stack source,
// such as Damage resolving a group relative to its own DamageSource, builds the
// resolver with newReferenceResolverWithSource so source-dependent references
// (the source permanent and the EquippedCreature/Other... selectors) bind to the
// caller's permanent, including an explicit nil that resolves to no source.
type referenceResolver struct {
	g   *game.Game
	obj *game.StackObject

	// controller overrides obj.Controller when controllerFixed is set. Static
	// continuous effects have a controller and source permanent but no resolving
	// stack object, so this avoids manufacturing one for every applicability
	// check.
	controller      game.PlayerID
	controllerFixed bool

	// source overrides the stack-derived source permanent when sourceFixed is
	// set. A nil source with sourceFixed true means "no source permanent".
	source      *game.Permanent
	sourceFixed bool

	// groupOfferMember, when set, resolves PlayerReferenceGroupOfferMember to the
	// player currently being offered an OptionalActorGroup instruction. The
	// effectResolver copies its live group-offer member into every reference
	// resolver it builds for group and player resolution, so a group scoped to the
	// acting player (PlayerControlledGroup(GroupOfferMemberReference(), ...) —
	// "each creature you control" under the Tempting-offer idiom) binds to the
	// controller for the base and reward resolutions and to each accepting opponent
	// for that opponent's own resolution. It is unset outside a group offer.
	groupOfferMember opt.V[game.PlayerID]
}

// withGroupOfferMember returns a copy of the resolver that resolves
// PlayerReferenceGroupOfferMember to member, so group and player references
// scoped to the acting player of an OptionalActorGroup instruction resolve while
// that member is being offered the effect.
func (r referenceResolver) withGroupOfferMember(member opt.V[game.PlayerID]) referenceResolver {
	r.groupOfferMember = member
	return r
}

func newReferenceResolver(g *game.Game, obj *game.StackObject) referenceResolver {
	return referenceResolver{g: g, obj: obj}
}

// targetSlot maps a reference's compile-time target-slot index to the position of
// that target within the resolving object's compacted Targets slice, skipping the
// slots that cast-branch-inactive gated specs reserved but chose no target for. It
// is the identity for every object without gated target specs, so ordinary target
// dereferencing is unchanged.
func (r referenceResolver) targetSlot(index int) int {
	return remapTargetSlot(r.g, r.obj, index)
}

// newReferenceResolverWithSource builds a resolver whose source permanent is
// fixed to source rather than derived from obj.SourceID. Passing a nil source
// fixes the basis to "no source permanent", preserving the legacy behavior of
// source-dependent selectors when the calling context has no source object.
func newReferenceResolverWithSource(g *game.Game, obj *game.StackObject, source *game.Permanent) referenceResolver {
	return referenceResolver{g: g, obj: obj, source: source, sourceFixed: true}
}

func newReferenceResolverWithControllerAndSource(g *game.Game, controller game.PlayerID, source *game.Permanent) referenceResolver {
	return referenceResolver{
		g:               g,
		controller:      controller,
		controllerFixed: true,
		source:          source,
		sourceFixed:     true,
	}
}

func (r referenceResolver) resolvingController() game.PlayerID {
	if r.controllerFixed {
		return r.controller
	}
	if r.obj != nil {
		return r.obj.Controller
	}
	return 0
}

type resolvedObjectReference struct {
	permanent            *game.Permanent
	snapshot             game.ObjectSnapshot
	stack                *game.StackObject
	stackController      game.PlayerID
	stackControllerKnown bool
}

func (r *resolvedObjectReference) controller(g *game.Game) (game.PlayerID, bool) {
	if r.permanent != nil {
		return effectiveController(g, r.permanent), true
	}
	if r.snapshot.ObjectID != 0 {
		return r.snapshot.Controller, true
	}
	if r.stack != nil {
		return r.stack.Controller, true
	}
	if r.stackControllerKnown {
		return r.stackController, true
	}
	return 0, false
}

func (r *resolvedObjectReference) owner(g *game.Game) (game.PlayerID, bool) {
	if r.permanent != nil {
		return r.permanent.Owner, true
	}
	if r.snapshot.ObjectID != 0 {
		return r.snapshot.Owner, true
	}
	if r.stack != nil {
		if r.stack.Kind != game.StackSpell || r.stack.Copy {
			return r.stack.Controller, true
		}
		if r.stack.SourceCardID != 0 {
			if card, ok := g.GetCardInstance(r.stack.SourceCardID); ok {
				return card.Owner, true
			}
		}
		if card, ok := g.GetCardInstance(r.stack.SourceID); ok {
			return card.Owner, true
		}
		return r.stack.Controller, true
	}
	return 0, false
}

// object resolves an ObjectReference to a live permanent or its last-known
// snapshot.
func (r referenceResolver) object(ref game.ObjectReference) (resolvedObjectReference, bool) {
	switch ref.Kind() {
	case game.ObjectReferenceTargetPermanent:
		objectID, ok := targetPermanentObjectID(r.g, r.obj, ref.TargetIndex())
		if !ok {
			return resolvedObjectReference{}, false
		}
		return resolvePermanentOrLastKnown(r.g, objectID)
	case game.ObjectReferenceTargetStackObject:
		if r.obj == nil {
			return resolvedObjectReference{}, false
		}
		objectID, ok := effectStackObjectID(r.g, r.obj, ref.TargetIndex())
		if !ok {
			controller, known := r.obj.TargetControllerLKI[ref.TargetIndex()]
			return resolvedObjectReference{
				stackController:      controller,
				stackControllerKnown: known,
			}, known
		}
		stackObject, ok := stackObjectByID(r.g, objectID)
		if ok {
			return resolvedObjectReference{stack: stackObject}, true
		}
		controller, ok := r.obj.TargetControllerLKI[ref.TargetIndex()]
		return resolvedObjectReference{
			stackController:      controller,
			stackControllerKnown: ok,
		}, ok
	case game.ObjectReferenceSourcePermanent:
		if r.sourceFixed {
			if r.source == nil || r.source.PhasedOut {
				return resolvedObjectReference{}, false
			}
			return resolvedObjectReference{permanent: r.source}, true
		}
		if r.obj == nil {
			return resolvedObjectReference{}, false
		}
		return resolveSourcePermanentOrLastKnown(r.g, r.obj.SourceID)
	case game.ObjectReferenceSourceCard:
		if r.obj == nil || r.obj.SourceCardID == 0 {
			return resolvedObjectReference{}, false
		}
		for _, permanent := range r.g.Battlefield {
			if permanent.CardInstanceID == r.obj.SourceCardID {
				return resolvedObjectReference{permanent: permanent}, true
			}
		}
		return resolvedObjectReference{}, false
	case game.ObjectReferenceSourceAttachedPermanent:
		permanent, ok := r.sourcePermanent()
		if !ok || !permanent.AttachedTo.Exists {
			return resolvedObjectReference{}, false
		}
		return resolvePermanentOrLastKnown(r.g, permanent.AttachedTo.Val)
	case game.ObjectReferenceTargetAttachedPermanent:
		objectID, ok := targetPermanentObjectID(r.g, r.obj, ref.TargetIndex())
		if !ok {
			return resolvedObjectReference{}, false
		}
		permanent, ok := permanentByObjectID(r.g, objectID)
		if !ok || !permanent.AttachedTo.Exists {
			return resolvedObjectReference{}, false
		}
		return resolvePermanentOrLastKnown(r.g, permanent.AttachedTo.Val)
	case game.ObjectReferenceLinkedObject:
		for _, linked := range linkedObjects(r.g, linkedObjectSourceKey(r.g, r.obj, ref.LinkID())) {
			if resolved, ok := resolvePermanentOrLastKnown(r.g, linked.ObjectID); ok {
				return resolved, true
			}
			// A card-only linked reference (ObjectID zero) names a card that was
			// never a battlefield permanent, such as a card exiled straight from a
			// graveyard (The Aesir Escape Valhalla). It carries no object snapshot,
			// so resolve it to a card snapshot the printed-characteristic readers
			// (mana value) consult through the card instance.
			if linked.ObjectID == 0 && linked.CardID != 0 {
				if _, ok := r.g.GetCardInstance(linked.CardID); ok {
					return resolvedObjectReference{snapshot: game.ObjectSnapshot{CardID: linked.CardID}}, true
				}
			}
			if linked.CardID != 0 {
				return resolvedObjectReference{snapshot: game.ObjectSnapshot{
					CardID: linked.CardID,
					Face:   game.FaceFront,
				}}, true
			}
		}
		return resolvedObjectReference{}, false
	case game.ObjectReferenceEventPermanent:
		if permanentID, ok := triggerEventPermanentID(r.obj); ok {
			return resolvePermanentOrLastKnown(r.g, permanentID)
		}
		return resolvedObjectReference{}, false
	case game.ObjectReferenceEventStackObject:
		if r.obj == nil || !r.obj.HasTriggerEvent || r.obj.TriggerEvent.StackObjectID == 0 {
			return resolvedObjectReference{}, false
		}
		stack, ok := stackObjectByID(r.g, r.obj.TriggerEvent.StackObjectID)
		if !ok {
			return resolvedObjectReference{}, false
		}
		return resolvedObjectReference{stack: stack}, true
	case game.ObjectReferenceEventRelatedPermanent:
		if r.obj != nil && r.obj.HasTriggerEvent && r.obj.TriggerEvent.RelatedPermanentID != 0 {
			return resolvePermanentOrLastKnown(r.g, r.obj.TriggerEvent.RelatedPermanentID)
		}
		return resolvedObjectReference{}, false
	case game.ObjectReferenceCapturedObject:
		if r.obj != nil && r.obj.CapturedObjectID != 0 {
			return resolvePermanentOrLastKnown(r.g, r.obj.CapturedObjectID)
		}
		return resolvedObjectReference{}, false
	case game.ObjectReferenceTargetObject:
		return r.targetObject(ref.TargetIndex())
	case game.ObjectReferenceTargetCard:
		cardID, ok := targetCardID(r.g, r.obj, ref.TargetIndex())
		if !ok {
			return resolvedObjectReference{}, false
		}
		if _, ok := r.g.GetCardInstance(cardID); !ok {
			return resolvedObjectReference{}, false
		}
		return resolvedObjectReference{snapshot: game.ObjectSnapshot{
			CardID: cardID,
			Face:   game.FaceFront,
		}}, true
	case game.ObjectReferenceSacrificedCost:
		if r.obj == nil || len(r.obj.SacrificedAsCostIDs) == 0 {
			return resolvedObjectReference{}, false
		}
		return resolvePermanentOrLastKnown(r.g, r.obj.SacrificedAsCostIDs[0])
	default:
		return resolvedObjectReference{}, false
	}
}

// targetObject resolves a kind-agnostic target reference to whichever permanent
// or stack object was chosen for the slot, delegating to the kind-specific
// resolution so a combined "spell or permanent" target binds either choice.
func (r referenceResolver) targetObject(index int) (resolvedObjectReference, bool) {
	slot := r.targetSlot(index)
	if r.obj == nil || slot < 0 || slot >= len(r.obj.Targets) {
		return resolvedObjectReference{}, false
	}
	switch r.obj.Targets[slot].Kind {
	case game.TargetPermanent:
		return r.object(game.TargetPermanentReference(index))
	case game.TargetStackObject:
		return r.object(game.TargetStackObjectReference(index))
	default:
		return resolvedObjectReference{}, false
	}
}

// player resolves a PlayerReference to a live player.
func (r referenceResolver) player(ref game.PlayerReference) (game.PlayerID, bool) {
	var playerID game.PlayerID
	var ok bool
	switch ref.Kind() {
	case game.PlayerReferenceController:
		playerID, ok = r.resolvingController(), true
	case game.PlayerReferenceTargetPlayer:
		playerID, ok = r.targetPlayer(ref.TargetIndex())
	case game.PlayerReferenceObjectController:
		playerID, ok = r.referencedObjectController(ref)
	case game.PlayerReferenceAffectedTargetController:
		playerID, ok = r.affectedTargetController(ref.TargetIndex())
	case game.PlayerReferenceObjectOwner:
		playerID, ok = r.referencedObjectOwner(ref)
	case game.PlayerReferenceEventPlayer:
		if r.obj.HasTriggerEvent {
			playerID, ok = triggeringEventPlayer(r.obj.TriggerEvent)
		}
	case game.PlayerReferenceCapturedTargetController:
		playerID, ok = r.obj.CapturedTargetControllerLKI[ref.TargetIndex()]
	case game.PlayerReferenceDefendingPlayer:
		if r.obj.HasTriggerEvent && defendingPlayerEvent(r.obj.TriggerEvent.Kind) {
			playerID, ok = r.obj.TriggerEvent.Player, true
		}
	case game.PlayerReferenceGiftRecipient:
		if r.obj.GiftPromised {
			playerID, ok = r.obj.GiftRecipient, true
		}
	case game.PlayerReferenceGroupOfferMember:
		if r.groupOfferMember.Exists {
			playerID, ok = r.groupOfferMember.Val, true
		}
	default:
		return 0, false
	}
	if !ok || !isPlayerAlive(r.g, playerID) {
		return 0, false
	}
	return playerID, true
}

// defendingPlayerEvent reports whether an event kind carries the defending
// player of an attack in Event.Player. The attacker-declared event sets it at
// declare-attackers; the became-blocked and became-unblocked events set it from
// the attacker's declared target at declare-blockers, so a "defending player ..."
// effect on any of these combat triggers resolves to the attacked player.
func defendingPlayerEvent(kind game.EventKind) bool {
	switch kind {
	case game.EventAttackerDeclared,
		game.EventAttackerBecameBlocked,
		game.EventAttackerBecameUnblocked:
		return true
	default:
		return false
	}
}

// triggerEventPermanentID resolves an ObjectReferenceEventPermanent ("it") to the
// permanent the trigger event is about. For most events this is the event's
// PermanentID (the permanent that entered, left, was damaged, attacked, or
// blocked). A damage event dealt to a player carries no damaged permanent, so
// the event permanent is instead the damage source (SourceObjectID); this lets
// "it" on a "Whenever ~ deals combat damage to a player, ..." trigger resolve to
// the damage source, as Cyclonus, the Saboteur's connive and Cyclonus,
// Cybertronian Fighter's convert require.
func triggerEventPermanentID(obj *game.StackObject) (id.ID, bool) {
	if obj == nil || !obj.HasTriggerEvent {
		return 0, false
	}
	event := obj.TriggerEvent
	if event.PermanentID != 0 {
		return event.PermanentID, true
	}
	if event.Kind == game.EventDamageDealt && event.SourceObjectID != 0 {
		return event.SourceObjectID, true
	}
	return 0, false
}

func triggeringEventPlayer(event game.Event) (game.PlayerID, bool) {
	switch event.Kind {
	case game.EventSpellCast, game.EventSpellCopied, game.EventPermanentTapped,
		game.EventAttackerDeclared:
		// EventAttackerDeclared records the attacking player in Controller and the
		// defending player in Player, so "that player" on an attack trigger
		// ("Whenever an opponent attacks you, ~ deals damage to that player.",
		// Emberwilde Captain) names the attacker, like the cast/tap events.
		return event.Controller, true
	case game.EventCardDrawn,
		game.EventCardDiscarded,
		game.EventCycled,
		game.EventPermanentSacrificed,
		game.EventScry,
		game.EventSurveil,
		game.EventAbilityActivated,
		game.EventBeginningOfStep,
		game.EventLifeGained,
		game.EventLifeLost,
		game.EventLibrarySearched,
		game.EventBecameMonarch,
		game.EventGotCityBlessing,
		game.EventCardPlayedFromExile,
		// EventManaProduced sets Player to the player who added the mana (the
		// mana ability's controller), so "that player adds an additional {U}"
		// (High Tide) names the player who tapped the Island for mana.
		game.EventManaProduced,
		// Every zone-change emission sets Player to the moved card's owner: the
		// leaving/entering permanent's Owner, the resolving stack object's card
		// Owner, or the player whose hand/library/graveyard/exile/command zone the
		// move runs through (discard, mill, and the generic move helpers all pass
		// that player as Player). Because a card is always put into its owner's
		// graveyard, hand, or library (CR 400.7, CR 404.2), Player names the
		// destination zone's owner for those transitions, so "that player" on
		// "Whenever a card is put into an opponent's graveyard ..." (Bloodchief
		// Ascension) resolves to the opponent whose graveyard received the card.
		game.EventZoneChanged:
		return event.Player, true
	case game.EventDamageDealt:
		return event.Player, event.DamageRecipient == game.DamageRecipientPlayer
	default:
		return 0, false
	}
}

// playerGroup resolves a PlayerGroupReference to alive players in stable player order.
func (r referenceResolver) playerGroup(ref game.PlayerGroupReference) []game.PlayerID {
	switch ref.Kind {
	case game.PlayerGroupReferenceOpponents:
		return aliveOpponents(r.g, r.resolvingController())
	case game.PlayerGroupReferenceAllPlayers:
		players := make([]game.PlayerID, 0, game.NumPlayers)
		for _, player := range r.g.Players {
			if !player.Eliminated {
				players = append(players, player.ID)
			}
		}
		return players
	case game.PlayerGroupReferenceTargetedPlayers:
		if r.obj == nil {
			return nil
		}
		players := make([]game.PlayerID, 0, len(r.obj.Targets))
		for _, target := range r.obj.Targets {
			if target.Kind == game.TargetPlayer && isPlayerAlive(r.g, target.PlayerID) {
				players = append(players, target.PlayerID)
			}
		}
		return players
	case game.PlayerGroupReferenceOpponentsAttackingTriggerPlayer:
		return r.opponentsAttackingTriggerPlayer()
	default:
		return nil
	}
}

// opponentsAttackingTriggerPlayer resolves the opponents of the resolving
// controller who have a creature attacking the player the resolving triggered
// ability's attack event was declared against ("Each opponent attacking that
// player does the same" — Curse of Opulence). "That player" is read from the
// trigger event, so it stays correct even if the source Aura has since left the
// battlefield. Only creatures attacking that player directly count; an attack on
// that player's planeswalker or battle is not an attack on the player (CR 508.1).
// Each opponent appears at most once no matter how many creatures they attack
// with, and the caller creates tokens in APNAP order.
func (r referenceResolver) opponentsAttackingTriggerPlayer() []game.PlayerID {
	if r.obj == nil || !r.obj.HasTriggerEvent || r.g.Combat == nil {
		return nil
	}
	attacked := r.obj.TriggerEvent.AttackTarget
	if !attacked.IsPlayerAttack() {
		return nil
	}
	controller := r.resolvingController()
	var members []game.PlayerID
	seen := make([]bool, game.NumPlayers)
	for _, declaration := range r.g.Combat.Attackers {
		if !declaration.Target.IsPlayerAttack() || declaration.Target.Player != attacked.Player {
			continue
		}
		attacker, ok := permanentByObjectID(r.g, declaration.Attacker)
		if !ok {
			continue
		}
		opponent := effectiveController(r.g, attacker)
		if opponent == controller || int(opponent) >= game.NumPlayers || seen[opponent] {
			continue
		}
		if !isPlayerAlive(r.g, opponent) {
			continue
		}
		seen[opponent] = true
		members = append(members, opponent)
	}
	return members
}

// permanentAt resolves a target slot to a permanent.
func (r referenceResolver) permanentAt(targetIndex int) (*game.Permanent, bool) {
	return effectPermanentTarget(r.g, r.obj, targetIndex)
}

// playerAt resolves a target slot to a player.
func (r referenceResolver) playerAt(targetIndex int) (game.PlayerID, bool) {
	return r.targetPlayer(targetIndex)
}

func (r referenceResolver) sourcePermanent() (*game.Permanent, bool) {
	if r.sourceFixed {
		return r.source, r.source != nil
	}
	return sourcePermanent(r.g, r.obj)
}

func (r referenceResolver) targetPlayer(targetIndex int) (game.PlayerID, bool) {
	slot := r.targetSlot(targetIndex)
	if slot < 0 || slot >= len(r.obj.Targets) {
		return 0, false
	}
	target := r.obj.Targets[slot]
	if target.Kind != game.TargetPlayer || !isPlayerAlive(r.g, target.PlayerID) {
		return 0, false
	}
	return target.PlayerID, true
}

// affectedTargetController resolves the player affected by a target slot: a
// player target resolves to that player, while any other target (a permanent,
// including a planeswalker) resolves to that object's controller. It backs the
// "that player or that permanent's controller" / "that creature's controller"
// payer and copier of the copy-chain family.
func (r referenceResolver) affectedTargetController(targetIndex int) (game.PlayerID, bool) {
	slot := r.targetSlot(targetIndex)
	if r.obj == nil || slot < 0 || slot >= len(r.obj.Targets) {
		return 0, false
	}
	if r.obj.Targets[slot].Kind == game.TargetPlayer {
		return r.targetPlayer(targetIndex)
	}
	resolved, ok := r.targetObject(targetIndex)
	if !ok {
		return 0, false
	}
	return resolved.controller(r.g)
}

func (r referenceResolver) referencedObjectController(ref game.PlayerReference) (game.PlayerID, bool) {
	object, ok := ref.Object()
	if !ok {
		return 0, false
	}
	resolved, ok := r.object(object)
	if !ok {
		return 0, false
	}
	return resolved.controller(r.g)
}

func (r referenceResolver) referencedObjectOwner(ref game.PlayerReference) (game.PlayerID, bool) {
	object, ok := ref.Object()
	if !ok {
		return 0, false
	}
	resolved, ok := r.object(object)
	if !ok {
		return 0, false
	}
	return resolved.owner(r.g)
}

// groupMembers enumerates the object IDs of the permanents that belong to a
// resolved GroupReference, preserving battlefield iteration order. It owns the
// candidate-domain enumeration and the object-reference exclusions that
// Selection deliberately keeps outside itself.
func (r referenceResolver) groupMembers(ref game.GroupReference) []id.ID {
	switch ref.Domain() {
	case game.GroupDomainAttachedObject:
		return r.attachedObjectGroupMembers(ref)
	case game.GroupDomainObjectControlled:
		return r.objectControlledGroupMembers(ref)
	case game.GroupDomainPlayerControlled:
		return r.playerControlledGroupMembers(ref)
	case game.GroupDomainBattlefield:
		return r.battlefieldGroupMembers(ref)
	case game.GroupDomainSameName:
		return r.sameNameGroupMembers(ref)
	case game.GroupDomainTriggeringAttackers:
		return r.triggeringAttackersGroupMembers(ref)
	case game.GroupDomainCapturedObjects:
		return r.capturedObjectsGroupMembers()
	default:
		return []id.ID{}
	}
}

// capturedObjectsGroupMembers enumerates the permanents a delayed trigger froze
// under its CapturedObjectGroup reference at schedule time, kept on the resolving
// stack object as CapturedObjectIDs. Members that have left the battlefield since
// capture are skipped ("Exile the tokens at end of combat.", the myriad keyword).
func (r referenceResolver) capturedObjectsGroupMembers() []id.ID {
	if r.obj == nil {
		return []id.ID{}
	}
	members := make([]id.ID, 0, len(r.obj.CapturedObjectIDs))
	for _, objectID := range r.obj.CapturedObjectIDs {
		if _, ok := permanentByObjectID(r.g, objectID); ok {
			members = append(members, objectID)
		}
	}
	return members
}

func (r referenceResolver) attachedObjectGroupMembers(ref game.GroupReference) []id.ID {
	anchor, ok := ref.Anchor()
	if !ok {
		return []id.ID{}
	}
	resolved, ok := r.object(anchor)
	if !ok || resolved.permanent == nil || !resolved.permanent.AttachedTo.Exists {
		return []id.ID{}
	}
	attachedID := resolved.permanent.AttachedTo.Val
	if _, ok := permanentByObjectID(r.g, attachedID); !ok {
		return []id.ID{}
	}
	return []id.ID{attachedID}
}

func (r referenceResolver) battlefieldGroupMembers(ref game.GroupReference) []id.ID {
	sel := ref.Selection()
	source, _ := r.sourcePermanent()
	if sel.ExcludeSource && source == nil {
		return []id.ID{}
	}
	excludedID := r.exclusionObjectID(ref)
	members := make([]id.ID, 0, len(r.g.Battlefield))
	for _, permanent := range r.g.Battlefield {
		if excludedID != 0 && permanent.ObjectID == excludedID {
			continue
		}
		if !r.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		members = append(members, permanent.ObjectID)
	}
	return members
}

// triggeringAttackersGroupMembers enumerates the creatures declared as attackers
// in the attack that triggered the resolving ability, narrowed by the group's
// Selection. It reconstructs the declared-attacker set from the resolving
// ability's trigger event: an EventAttackerDeclared trigger coalesces its
// simultaneous batch into a single trigger and retains the first matching event
// (game.TriggerPattern.OneOrMore), so the attackers are the permanents named by
// every EventAttackerDeclared event sharing the trigger event's SimultaneousID
// (or the single triggering event when unbatched). Binding the declared
// attackers rather than re-querying the board excludes a creature that started
// attacking after the declaration and still includes a declared attacker that
// left combat before resolution. It fails closed when the resolving ability has
// no attacker-declared trigger event.
func (r referenceResolver) triggeringAttackersGroupMembers(ref game.GroupReference) []id.ID {
	if r.obj == nil || !r.obj.HasTriggerEvent || r.obj.TriggerEvent.Kind != game.EventAttackerDeclared {
		return []id.ID{}
	}
	sel := ref.Selection()
	source, _ := r.sourcePermanent()
	defenderFilter := ref.AttackedDefenderFilter()
	members := make([]id.ID, 0)
	for _, objectID := range triggeringAttackerObjectIDs(r.g, r.obj.TriggerEvent) {
		permanent, ok := permanentByObjectID(r.g, objectID)
		if !ok {
			continue
		}
		if !r.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		if defenderFilter != game.TriggerControllerAny {
			defendingPlayer, ok := triggeringAttackerDefendingPlayer(r.g, r.obj.TriggerEvent, objectID)
			if !ok || !triggerControllerMatches(r.resolvingController(), defenderFilter, defendingPlayer) {
				continue
			}
		}
		members = append(members, objectID)
	}
	return members
}

// triggeringAttackerDefendingPlayer returns the player an attacker was declared
// against in the trigger's simultaneous batch, read from the stored
// EventAttackerDeclared events (which record the defending player in Player, even
// when a planeswalker or battle was the direct target). Reading the declaration
// rather than re-deriving it from live combat keeps a declared attacker that has
// since left combat classified by its original defender, matching the
// declared-batch snapshot the group binds. It reports false when no declaration
// for the attacker is found.
func triggeringAttackerDefendingPlayer(g *game.Game, trigger game.Event, attackerObjectID id.ID) (game.PlayerID, bool) {
	if trigger.SimultaneousID == 0 {
		if trigger.PermanentID == attackerObjectID {
			return trigger.Player, true
		}
		return 0, false
	}
	for i := range g.Events {
		event := &g.Events[i]
		if event.Kind == game.EventAttackerDeclared &&
			event.SimultaneousID == trigger.SimultaneousID &&
			event.PermanentID == attackerObjectID {
			return event.Player, true
		}
	}
	return 0, false
}

// triggeringAttackerObjectIDs returns the object IDs of the creatures declared
// as attackers in the trigger's simultaneous batch, in declaration order. An
// unbatched trigger (SimultaneousID zero) yields its single attacker.
func triggeringAttackerObjectIDs(g *game.Game, trigger game.Event) []id.ID {
	if trigger.SimultaneousID == 0 {
		if trigger.PermanentID == 0 {
			return nil
		}
		return []id.ID{trigger.PermanentID}
	}
	var ids []id.ID
	seen := make(map[id.ID]bool)
	for _, event := range g.Events {
		if event.Kind != game.EventAttackerDeclared ||
			event.SimultaneousID != trigger.SimultaneousID ||
			event.PermanentID == 0 ||
			seen[event.PermanentID] {
			continue
		}
		seen[event.PermanentID] = true
		ids = append(ids, event.PermanentID)
	}
	return ids
}

// sameNameGroupMembers enumerates every battlefield permanent whose name equals
// the anchor object's name and satisfies the group's Selection, the "<target>
// and all other <group> with the same name as that permanent" wording (Maelstrom
// Pulse, the Echoing cycle). The anchor permanent is included because it shares
// its own name. A missing anchor, an anchor with no available name (a face-down
// permanent or a token without a card def), or a permanent with no available
// name fails closed, so an unnamed object never matches.
func (r referenceResolver) sameNameGroupMembers(ref game.GroupReference) []id.ID {
	anchor, ok := ref.Anchor()
	if !ok {
		return []id.ID{}
	}
	resolved, ok := r.object(anchor)
	if !ok || resolved.permanent == nil {
		return []id.ID{}
	}
	anchorName, ok := permanentNameForSameNameGroup(r.g, resolved.permanent)
	if !ok {
		return []id.ID{}
	}
	sel := ref.Selection()
	source, _ := r.sourcePermanent()
	if sel.ExcludeSource && source == nil {
		return []id.ID{}
	}
	members := make([]id.ID, 0, len(r.g.Battlefield))
	for _, permanent := range r.g.Battlefield {
		name, ok := permanentNameForSameNameGroup(r.g, permanent)
		if !ok || name != anchorName {
			continue
		}
		if !r.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		members = append(members, permanent.ObjectID)
	}
	return members
}

// permanentNameForSameNameGroup resolves a permanent's effective name for
// same-name group matching, reporting false when no name is available (a
// face-down permanent or a permanent without a card def) so an unnamed object
// neither anchors nor joins the group.
func permanentNameForSameNameGroup(g *game.Game, permanent *game.Permanent) (string, bool) {
	if permanent == nil || permanent.FaceDown {
		return "", false
	}
	if _, ok := permanentCardDef(g, permanent); !ok {
		return "", false
	}
	return permanentEffectiveName(g, permanent), true
}

func (r referenceResolver) objectControlledGroupMembers(ref game.GroupReference) []id.ID {
	anchor, ok := ref.Anchor()
	if !ok {
		return []id.ID{}
	}
	resolved, ok := r.object(anchor)
	if !ok {
		return []id.ID{}
	}
	controller, ok := resolved.controller(r.g)
	if !ok {
		return []id.ID{}
	}
	sel := ref.Selection()
	source, _ := r.sourcePermanent()
	if sel.ExcludeSource && source == nil {
		return []id.ID{}
	}
	excludedID := r.exclusionObjectID(ref)
	members := make([]id.ID, 0, len(r.g.Battlefield))
	for _, permanent := range r.g.Battlefield {
		if excludedID != 0 && permanent.ObjectID == excludedID {
			continue
		}
		if effectiveController(r.g, permanent) != controller {
			continue
		}
		if !r.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		members = append(members, permanent.ObjectID)
	}
	return members
}

// playerControlledGroupMembers enumerates the battlefield permanents controlled
// by the player named by the group's anchor player reference and satisfying its
// Selection, such as every creature a targeted player controls. It mirrors
// objectControlledGroupMembers but resolves the controlling player directly from
// the anchor player reference rather than from an anchor object's controller.
func (r referenceResolver) playerControlledGroupMembers(ref game.GroupReference) []id.ID {
	anchor, ok := ref.PlayerAnchor()
	if !ok {
		return []id.ID{}
	}
	controller, ok := r.player(anchor)
	if !ok {
		return []id.ID{}
	}
	sel := ref.Selection()
	source, _ := r.sourcePermanent()
	if sel.ExcludeSource && source == nil {
		return []id.ID{}
	}
	excludedID := r.exclusionObjectID(ref)
	members := make([]id.ID, 0, len(r.g.Battlefield))
	for _, permanent := range r.g.Battlefield {
		if excludedID != 0 && permanent.ObjectID == excludedID {
			continue
		}
		if effectiveController(r.g, permanent) != controller {
			continue
		}
		if !r.permanentMatchesGroupSelection(&sel, source, permanent) {
			continue
		}
		members = append(members, permanent.ObjectID)
	}
	return members
}

// permanentMatchesGroupSelection reports whether permanent satisfies the group's
// Selection. Callers hoist the source-exclusion divergence (an ExcludeSource
// group with no source permanent matches nothing) out of the enumeration loop.
func (r referenceResolver) permanentMatchesGroupSelection(sel *game.Selection, source, permanent *game.Permanent) bool {
	values := effectivePermanentValues(r.g, permanent)
	subject := selectionSubject{
		kind:              subjectPermanent,
		g:                 r.g,
		permanent:         permanent,
		values:            values,
		viewer:            r.resolvingController(),
		resolutionChoices: r.obj.ResolutionChoices,
		obj:               r.obj,
	}
	if sel.Controller != game.ControllerAny {
		subject.controller = effectiveController(r.g, permanent)
	}
	if source != nil {
		subject.sourceObjectID = source.ObjectID
	}
	return matchSelection(&subject, sel)
}

// exclusionObjectID resolves the group's excluded object to its identity object
// ID, returning 0 when the group has no exclusion or the excluded object cannot
// be identified. It reads the reference's identity even when the permanent has
// left the battlefield, matching the legacy "except target" exclusion.
func (r referenceResolver) exclusionObjectID(ref game.GroupReference) id.ID {
	exclude, ok := ref.Exclusion()
	if !ok {
		return 0
	}
	objectID, _ := r.objectIdentityID(exclude)
	return objectID
}

func (r referenceResolver) objectIdentityID(ref game.ObjectReference) (id.ID, bool) {
	switch ref.Kind() {
	case game.ObjectReferenceTargetPermanent:
		return targetPermanentObjectID(r.g, r.obj, ref.TargetIndex())
	case game.ObjectReferenceSourcePermanent:
		if r.obj == nil || r.obj.SourceID == 0 {
			return 0, false
		}
		return r.obj.SourceID, true
	case game.ObjectReferenceEventPermanent:
		return triggerEventPermanentID(r.obj)
	case game.ObjectReferenceEventRelatedPermanent:
		if r.obj != nil && r.obj.HasTriggerEvent && r.obj.TriggerEvent.RelatedPermanentID != 0 {
			return r.obj.TriggerEvent.RelatedPermanentID, true
		}
		return 0, false
	case game.ObjectReferenceCapturedObject:
		if r.obj != nil && r.obj.CapturedObjectID != 0 {
			return r.obj.CapturedObjectID, true
		}
		return 0, false
	case game.ObjectReferenceSourceAttachedPermanent:
		permanent, ok := r.sourcePermanent()
		if !ok || !permanent.AttachedTo.Exists {
			return 0, false
		}
		return permanent.AttachedTo.Val, true
	case game.ObjectReferenceTargetAttachedPermanent:
		objectID, ok := targetPermanentObjectID(r.g, r.obj, ref.TargetIndex())
		if !ok {
			return 0, false
		}
		permanent, ok := permanentByObjectID(r.g, objectID)
		if !ok || !permanent.AttachedTo.Exists {
			return 0, false
		}
		return permanent.AttachedTo.Val, true
	case game.ObjectReferenceLinkedObject:
		for _, linked := range linkedObjects(r.g, linkedObjectSourceKey(r.g, r.obj, ref.LinkID())) {
			if linked.ObjectID != 0 {
				return linked.ObjectID, true
			}
		}
		return 0, false
	default:
		return 0, false
	}
}

// resolveObjectReference is a thin adapter over the referenceResolver module.
func resolveObjectReference(g *game.Game, obj *game.StackObject, ref game.ObjectReference) (resolvedObjectReference, bool) {
	return newReferenceResolver(g, obj).object(ref)
}

// resolvePlayerReference is a thin adapter over the referenceResolver module.
func resolvePlayerReference(g *game.Game, obj *game.StackObject, ref game.PlayerReference) (game.PlayerID, bool) {
	return newReferenceResolver(g, obj).player(ref)
}

func resolvePermanentOrLastKnown(g *game.Game, objectID id.ID) (resolvedObjectReference, bool) {
	if permanent, ok := permanentByObjectID(g, objectID); ok {
		return resolvedObjectReference{permanent: permanent}, true
	}
	if snapshot, ok := lastKnownObject(g, objectID); ok {
		return resolvedObjectReference{snapshot: snapshot}, true
	}
	return resolvedObjectReference{}, false
}

func resolveSourcePermanentOrLastKnown(g *game.Game, objectID id.ID) (resolvedObjectReference, bool) {
	if permanent, ok := permanentByObjectID(g, objectID); ok {
		if !permanent.PhasedOut {
			return resolvedObjectReference{permanent: permanent}, true
		}
		if snapshot, ok := lastKnownObject(g, objectID); ok {
			return resolvedObjectReference{snapshot: snapshot}, true
		}
		snapshot := snapshotPermanent(g, permanent, zone.Battlefield)
		return resolvedObjectReference{snapshot: snapshot}, true
	}
	if snapshot, ok := lastKnownObject(g, objectID); ok {
		return resolvedObjectReference{snapshot: snapshot}, true
	}
	return resolvedObjectReference{}, false
}

func targetPermanentObjectID(g *game.Game, obj *game.StackObject, targetIndex int) (id.ID, bool) {
	if obj == nil {
		return 0, false
	}
	targetIndex = remapTargetSlot(g, obj, targetIndex)
	if targetIndex < 0 || targetIndex >= len(obj.Targets) {
		return 0, false
	}
	target := obj.Targets[targetIndex]
	if target.Kind != game.TargetPermanent {
		return 0, false
	}
	return target.PermanentID, true
}

// targetCardID returns the chosen card ID for a card target slot, such as a
// creature card selected in a graveyard. It backs ObjectReferenceTargetCard,
// which names that card as the blueprint of a copy token.
func targetCardID(g *game.Game, obj *game.StackObject, targetIndex int) (id.ID, bool) {
	if obj == nil {
		return 0, false
	}
	targetIndex = remapTargetSlot(g, obj, targetIndex)
	if targetIndex < 0 || targetIndex >= len(obj.Targets) {
		return 0, false
	}
	target := obj.Targets[targetIndex]
	if target.Kind != game.TargetCard {
		return 0, false
	}
	return target.CardID, true
}
