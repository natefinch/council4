package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func (r *effectResolver) damageSource(source game.ObjectReference) (effectDamageSource, bool) {
	if source.Kind() == game.ObjectReferenceNone {
		sourceID, sourceObjectID := damageSourceIDs(r.game, r.obj)
		return effectDamageSource{
			sourceID:       sourceID,
			sourceObjectID: sourceObjectID,
			controller:     r.obj.Controller,
		}, true
	}
	resolved, ok := resolveObjectReference(r.game, r.obj, source)
	if !ok {
		return effectDamageSource{}, false
	}
	if resolved.permanent == nil {
		if resolved.snapshot.ObjectID == 0 {
			return effectDamageSource{}, false
		}
		return effectDamageSource{
			sourceID:       resolved.snapshot.CardID,
			sourceObjectID: resolved.snapshot.ObjectID,
			controller:     resolved.snapshot.Controller,
			deathtouch:     slices.Contains(resolved.snapshot.Keywords, game.Deathtouch),
			lifelink:       slices.Contains(resolved.snapshot.Keywords, game.Lifelink),
		}, true
	}
	return effectDamageSource{
		sourceID:       resolved.permanent.CardInstanceID,
		sourceObjectID: resolved.permanent.ObjectID,
		controller:     effectiveController(r.game, resolved.permanent),
		permanent:      resolved.permanent,
		deathtouch:     hasKeyword(r.game, resolved.permanent, game.Deathtouch),
		lifelink:       hasKeyword(r.game, resolved.permanent, game.Lifelink),
	}, true
}

func handleDamage(r *effectResolver, prim game.Damage) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	sourceRef := game.ObjectReference{}
	if prim.DamageSource.Exists {
		sourceRef = prim.DamageSource.Val
	}
	source, ok := r.damageSource(sourceRef)
	if !ok {
		return res
	}
	if prim.Divided {
		return r.damageDivided(res, source, prim)
	}
	if object, ok := prim.Recipient.ObjectReference(); ok {
		return r.damageReferencedPermanent(res, source, prim.ResultAmountKind, object)
	}
	if player, ok := prim.Recipient.PlayerReference(); ok {
		return r.damageReferencedPlayer(res, source, prim.ResultAmountKind, player)
	}
	if player, ok := prim.Recipient.AnyTargetPlayerReference(); ok {
		if resolvedPlayer, playerOK := r.resolvePlayer(player); playerOK {
			dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, resolvedPlayer, res.amount, false)
			applyDamageSourceLifelink(r.game, source, dealt)
			res.amount = typedDamageResultAmount(prim.ResultAmountKind, dealt, 0)
			res.succeeded = dealt > 0
			return res
		}
	}
	if object, ok := prim.Recipient.AnyTargetObjectReference(); ok {
		return r.damageReferencedPermanent(res, source, prim.ResultAmountKind, object)
	}
	if group, ok := prim.Recipient.GroupReference(); ok {
		return r.damageSelectedPermanents(res, source, group)
	}
	if group, ok := prim.Recipient.PlayerGroupReference(); ok {
		for _, playerID := range r.playerGroupMembers(group) {
			dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, playerID, res.amount, false)
			applyDamageSourceLifelink(r.game, source, dealt)
			res.succeeded = dealt > 0 || res.succeeded
		}
	}
	return res
}

// dividedDamageTarget pairs a chosen runtime target with a stable label for the
// allocation choice prompt. legal records whether the target is still a legal
// recipient at resolution; an illegal target keeps its share of the division but
// the damage is simply not dealt (see damageDivided).
type dividedDamageTarget struct {
	target game.Target
	label  string
	legal  bool
}

// damageDivided splits a fixed total among the targets chosen for the
// recipient's target spec. The division covers every target chosen for the spec,
// each receiving at least one so the allocations sum to the total (CR 601.2d).
//
// CR 601.2d locks the division in over the targets chosen when the spell or
// ability was put on the stack; a target that has since become illegal is no
// longer dealt damage, and its assigned amount is lost rather than redistributed
// to the remaining targets. The allocation therefore spans every originally
// chosen target (legal or not), and only the still-legal targets are dealt their
// share.
func (r *effectResolver) damageDivided(res effectResolved, source effectDamageSource, prim game.Damage) effectResolved {
	object, ok := prim.Recipient.AnyTargetObjectReference()
	if !ok {
		return res
	}
	specIndex := object.TargetIndex()
	targets := r.dividedTargets(specIndex)
	if len(targets) == 0 {
		return res
	}
	allocations := r.allocateDividedDamage(res.amount, targets)
	dealtAny := false
	for i, entry := range targets {
		amount := allocations[i]
		// An illegal target keeps its share of the division but is dealt no
		// damage; the amount is lost, never redistributed (CR 601.2d).
		if amount <= 0 || !entry.legal {
			continue
		}
		switch entry.target.Kind {
		case game.TargetPermanent:
			permanent, found := permanentByObjectID(r.game, entry.target.PermanentID)
			if !found {
				continue
			}
			dealt := dealPermanentDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, permanent, amount, false)
			applyDamageSourceKeywordEffects(r.game, source, permanent, dealt)
			dealtAny = dealtAny || dealt > 0
		case game.TargetPlayer:
			dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, entry.target.PlayerID, amount, false)
			applyDamageSourceLifelink(r.game, source, dealt)
			dealtAny = dealtAny || dealt > 0
		default:
			continue
		}
	}
	res.amount = typedDamageResultAmount(prim.ResultAmountKind, res.amount, 0)
	res.succeeded = dealtAny
	return res
}

// dividedTargets returns every target chosen for the spec at specIndex, using
// TargetCounts to locate the spec's slice of obj.Targets. When target counts are
// unavailable it falls back to every chosen target, which matches the single-spec
// divided-damage shape the compiler emits.
//
// Targets that have become illegal since announcement are included with
// legal=false so they still hold their share of the division; damageDivided
// drops that share rather than dealing or redistributing it (CR 601.2d).
func (r *effectResolver) dividedTargets(specIndex int) []dividedDamageTarget {
	all := r.obj.Targets
	start, end := 0, len(all)
	if counts := r.obj.TargetCounts; specIndex >= 0 && specIndex < len(counts) {
		start = 0
		for i := range specIndex {
			start += counts[i]
		}
		end = start + counts[specIndex]
	}
	if start < 0 || end > len(all) || start > end {
		return nil
	}
	entries := make([]dividedDamageTarget, 0, end-start)
	for i := start; i < end; i++ {
		target := all[i]
		legal := false
		switch target.Kind {
		case game.TargetPermanent:
			permanent, found := permanentByObjectID(r.game, target.PermanentID)
			legal = found && activeBattlefieldPermanent(permanent)
		case game.TargetPlayer:
			legal = isPlayerAlive(r.game, target.PlayerID)
		default:
			// Only permanents and players are valid divided-damage recipients;
			// any other chosen-target kind is excluded from the division.
			continue
		}
		entries = append(entries, dividedDamageTarget{target: target, label: dividedTargetLabel(r.game, target), legal: legal})
	}
	return entries
}

func dividedTargetLabel(g *game.Game, target game.Target) string {
	if target.Kind == game.TargetPermanent {
		if permanent, ok := permanentByObjectID(g, target.PermanentID); ok {
			return permanentChoiceLabel(g, permanent)
		}
		return "permanent"
	}
	return "player"
}

// allocateDividedDamage asks the controller to split total among every target
// chosen for the spec, returning one allocation per target. Each target receives
// at least one and the allocations sum to total, including targets that have
// since become illegal (whose share damageDivided then drops). The
// ChoiceDamageAllocation response lists option indices with repetition, where the
// count of an index is its allocation.
func (r *effectResolver) allocateDividedDamage(total int, targets []dividedDamageTarget) []int {
	n := len(targets)
	allocations := make([]int, n)
	if total < n {
		// Not enough total to give each target one; this should be prevented at
		// announcement, but stay defensive and give one each up to the total.
		for i := range total {
			allocations[i] = 1
		}
		return allocations
	}
	options := make([]game.ChoiceOption, n)
	for i, entry := range targets {
		options[i] = game.ChoiceOption{Index: i, Label: entry.label}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceDamageAllocation,
		Player:           stackObjectController(r.obj),
		Prompt:           "Divide damage among the chosen targets.",
		Options:          options,
		MinChoices:       total,
		MaxChoices:       total,
		DefaultSelection: defaultDividedAllocation(total, n),
	}
	selected := r.engine.chooseChoice(r.game, r.agents, request, r.log)
	for _, index := range selected {
		if index >= 0 && index < n {
			allocations[index]++
		}
	}
	return allocations
}

// defaultDividedAllocation gives one to each of the first n-1 targets and the
// remainder to the last, expressed as a repeated-index selection.
func defaultDividedAllocation(total, n int) []int {
	if n <= 0 || total < n {
		return nil
	}
	selected := make([]int, 0, total)
	for i := 0; i < n-1; i++ {
		selected = append(selected, i)
	}
	for i := 0; i < total-(n-1); i++ {
		selected = append(selected, n-1)
	}
	return selected
}

func (r *effectResolver) damageReferencedPlayer(res effectResolved, source effectDamageSource, resultKind game.EffectResultAmountKind, player game.PlayerReference) effectResolved {
	playerID, ok := r.resolvePlayer(player)
	if !ok {
		return res
	}
	dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, playerID, res.amount, false)
	applyDamageSourceLifelink(r.game, source, dealt)
	res.amount = typedDamageResultAmount(resultKind, dealt, 0)
	res.succeeded = dealt > 0
	return res
}

func (r *effectResolver) damageReferencedPermanent(res effectResolved, source effectDamageSource, resultKind game.EffectResultAmountKind, object game.ObjectReference) effectResolved {
	permanent, ok := r.resolveObject(object)
	if !ok {
		return res
	}
	lethalRemaining := lethalDamageRemaining(r.game, permanent)
	if source.deathtouch {
		lethalRemaining = 1
		if permanent.MarkedDeathtouchDamage {
			lethalRemaining = 0
		}
	} else if source.permanent != nil {
		lethalRemaining = lethalDamageRemainingFromSource(r.game, source.permanent, permanent)
	}
	dealt := dealPermanentDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, permanent, res.amount, false)
	applyDamageSourceKeywordEffects(r.game, source, permanent, dealt)
	res.excessDamage = max(0, dealt-lethalRemaining)
	res.amount = typedDamageResultAmount(resultKind, dealt, res.excessDamage)
	res.succeeded = dealt > 0 && (resultKind != game.EffectResultAmountExcessDamage || res.excessDamage > 0)
	return res
}

func (r *effectResolver) damageSelectedPermanents(res effectResolved, source effectDamageSource, group game.GroupReference) effectResolved {
	for _, permanent := range r.groupPermanentsWithSource(group, source.permanent) {
		dealt := dealPermanentDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, permanent, res.amount, false)
		applyDamageSourceKeywordEffects(r.game, source, permanent, dealt)
		res.succeeded = dealt > 0 || res.succeeded
	}
	return res
}

func handleGroupSourceDamage(r *effectResolver, prim game.GroupSourceDamage) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	for _, permanent := range r.groupPermanents(prim.Group) {
		controller := effectiveController(r.game, permanent)
		recipient := controller
		if prim.ToOwner {
			recipient = permanent.Owner
		}
		dealt := dealPlayerDamage(r.game, permanent.CardInstanceID, permanent.ObjectID, controller, recipient, res.amount, false)
		res.succeeded = dealt > 0 || res.succeeded
	}
	return res
}

func typedDamageResultAmount(kind game.EffectResultAmountKind, dealt, excess int) int {
	if kind == game.EffectResultAmountExcessDamage {
		return excess
	}
	return dealt
}

func handleFight(r *effectResolver, prim game.Fight) effectResolved {
	first, firstOK := r.resolveObject(prim.Object)
	second, secondOK := r.resolveObject(prim.RelatedObject)
	if !firstOK || !secondOK || first.ObjectID == second.ObjectID ||
		!permanentHasType(r.game, first, types.Creature) || !permanentHasType(r.game, second, types.Creature) {
		return effectResolved{accepted: true}
	}
	resolveFightPermanents(r.game, first, second)
	return effectResolved{accepted: true, succeeded: true}
}

func handlePreventDamage(r *effectResolver, prim game.PreventDamage) effectResolved {
	res := effectResolved{accepted: true}
	if !prim.All {
		res.amount = r.quantity(prim.Amount)
	}
	res.succeeded = createPreventionShield(r.game, r.obj, res.amount, prim, game.DurationUntilEndOfTurn)
	return res
}
